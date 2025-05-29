package wal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/ethereum/go-ethereum/common"
)

type BlockWal struct {
	BlockInfo          BlockInfo
	OperationsIterator BlockOperationsIterator
}

func NewIterator(
	ctx context.Context,
	walDir string,
	nextBlockNumber uint64,
	prevBlockHash common.Hash,
	waitForNewBlocks bool,
) func(yield func(blockWal BlockWal, err error) bool) {

	blockNumber := nextBlockNumber

	return func(yield func(blockWal BlockWal, err error) bool) {

		for ctx.Err() == nil {

			filename := filepath.Join(walDir, BlockNumberToFilename(blockNumber))

			bi, operationsIterator, err := NewBlockOperationsIterator(ctx, filename)

			if errors.Is(err, os.ErrNotExist) {
				if !waitForNewBlocks {
					return
				}
				bo := backoff.WithContext(backoff.NewConstantBackOff(time.Second), ctx)
				backoff.Retry(func() error {
					_, err := os.Stat(filename)
					return err
				}, bo)

				continue
			}

			if err != nil {
				if !yield(BlockWal{}, fmt.Errorf("failed to create block operations iterator: %w", err)) {
					return
				}
			}

			if err != nil {
				if !yield(BlockWal{}, fmt.Errorf("failed to decode block: %w", err)) {
					return
				}
			}

			if bi.Number != blockNumber {
				if !yield(BlockWal{}, fmt.Errorf("block number mismatch: expected %d, got %d", blockNumber, bi.Number)) {
					return
				}
			}

			if bi.ParentHash != prevBlockHash {
				if !yield(BlockWal{}, fmt.Errorf("block hash mismatch: expected %s, got %s", prevBlockHash.Hex(), bi.Hash.Hex())) {
					return
				}
			}

			bw := BlockWal{
				BlockInfo:          bi,
				OperationsIterator: operationsIterator,
			}

			if !yield(bw, nil) {
				return
			}

			blockNumber = bi.Number + 1
			prevBlockHash = bi.Hash
		}
	}

}
