package eth

import (
	"context"
	"fmt"
	"math/big"

	sqlitestore "github.com/Arkiv-Network/sqlite-bitmap-store"
	"github.com/ethereum/go-ethereum/arkiv/storageaccounting"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
)

type arkivAPI struct {
	eth   *Ethereum
	store *sqlitestore.SQLiteStore
}

func NewArkivAPI(eth *Ethereum, store *sqlitestore.SQLiteStore) (*arkivAPI, error) {
	return &arkivAPI{
		eth:   eth,
		store: store,
	}, nil
}

func (api *arkivAPI) Query(
	ctx context.Context,
	req string,
	op *sqlitestore.Options,
) (*sqlitestore.QueryResponse, error) {

	lastBlock := api.eth.blockchain.CurrentHeader().Number.Uint64()

	log.Info("api", "last_block", lastBlock)

	if op == nil {
		op = &sqlitestore.Options{}
	}
	if op.AtBlock == nil {
		op.AtBlock = &lastBlock
	}

	response, err := api.store.QueryEntities(ctx, req, op)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}

	return response, nil
}

// GetEntityCount returns the total number of entities in the storage.
func (api *arkivAPI) GetEntityCount(ctx context.Context) (uint64, error) {

	count, err := api.store.GetNumberOfEntities(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get entity count: %w", err)
	}
	return count, nil

}

func (api *arkivAPI) GetNumberOfUsedSlots() (*hexutil.Big, error) {
	header := api.eth.blockchain.CurrentBlock()
	stateDB, err := api.eth.BlockChain().StateAt(header.Root)
	if err != nil {
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	counter := storageaccounting.GetNumberOfUsedSlots(stateDB)
	counterAsBigInt := big.NewInt(0)
	counter.IntoBig(&counterAsBigInt)
	return (*hexutil.Big)(counterAsBigInt), nil
}

type BlockTiming struct {
	CurrentBlock     uint64 `json:"current_block"`
	CurrentBlockTime uint64 `json:"current_block_time"`
	BlockDuration    uint64 `json:"duration"`
}

func (api *arkivAPI) GetBlockTiming(ctx context.Context) (*BlockTiming, error) {
	header := api.eth.blockchain.CurrentHeader()
	previousHeader := api.eth.blockchain.GetHeaderByHash(header.ParentHash)
	if previousHeader == nil {
		return nil, fmt.Errorf("failed to get previous header")
	}

	return &BlockTiming{
		CurrentBlock:     header.Number.Uint64(),
		CurrentBlockTime: header.Time,
		BlockDuration:    header.Time - previousHeader.Time,
	}, nil
}
