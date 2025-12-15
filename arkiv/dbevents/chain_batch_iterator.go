package dbevents

import (
	"sync"

	arkivevents "github.com/Arkiv-Network/arkiv-events"
	"github.com/Arkiv-Network/arkiv-events/events"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

func NewChainBatchIterator(db ethdb.Database, lastBlock uint64) (
	arkivevents.BatchIterator,
	func(cc *params.ChainConfig, block *types.Block) error,
) {

	mu := &sync.Mutex{}
	cond := sync.NewCond(mu)
	var block *types.Block

	var chainConfig *params.ChainConfig

	onNewHead := func(cc *params.ChainConfig, bl *types.Block) error {
		mu.Lock()
		block = bl
		chainConfig = cc
		cond.Signal()
		mu.Unlock()
		log.Info("Arkiv new head", "number", bl.Number, "hash", bl.Hash())
		return nil
	}

	batchIterator := arkivevents.BatchIterator(
		func(yield func(arkivevents.BatchOrError) bool) {

			for {

				batch := arkivevents.BatchOrError{
					Batch: events.BlockBatch{},
					Error: nil,
				}

				func() {
					mu.Lock()

					for block == nil {
						cond.Wait()
					}
					newBlockNumber := block.NumberU64()

					block = nil

					mu.Unlock()

					log.Info("Arkiv new head", "number", newBlockNumber)

					if newBlockNumber <= lastBlock {
						return
					}

					batchSize := min(100, (newBlockNumber - lastBlock))

					log.Info("Arkiv reading batch", "size", batchSize)

					for i := range batchSize {
						log.Info("Arkiv reading block", "number", lastBlock+i)

						blockNumber := lastBlock + i + 1

						hash := rawdb.ReadCanonicalHash(db, blockNumber)
						if hash == (common.Hash{}) {
							log.Warn("Canonical hash not found", "number", blockNumber)
							return
						}
						bl := rawdb.ReadBlock(db, hash, blockNumber)

						receiepts := rawdb.ReadReceipts(db, hash, bl.NumberU64(), bl.Time(), chainConfig)

						if receiepts == nil {
							log.Warn("receipts not found for block", "number", blockNumber, "hash", hash)
							return
						}

						block := rawdb.ReadBlock(db, bl.Hash(), bl.NumberU64())
						if block == nil {
							log.Warn("block not found for block", "number", blockNumber, "hash", hash)
							return
						}

						batchBlock, err := blockToEvents(block, receiepts)
						if err != nil {
							log.Error("failed to convert block to events", "number", blockNumber, "hash", hash, "error", err)
							return
						}

						batch.Batch.Blocks = append(batch.Batch.Blocks, *batchBlock)

					}

				}()

				if len(batch.Batch.Blocks) == 0 {
					continue
				}

				log.Info("yielding batch", "from", batch.Batch.Blocks[0].Number, "to", batch.Batch.Blocks[len(batch.Batch.Blocks)-1].Number)

				lastBlock = batch.Batch.Blocks[len(batch.Batch.Blocks)-1].Number

				if !yield(batch) {
					return
				}
			}

		},
	)

	return batchIterator, onNewHead
}
