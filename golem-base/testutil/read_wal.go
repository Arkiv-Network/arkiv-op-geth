package testutil

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/golem-base/wal"
)

func (w *World) ReadWAL(ctx context.Context) ([]wal.Operation, error) {

	client := w.GethInstance.ETHClient
	genesisHeader, err := client.HeaderByNumber(ctx, big.NewInt(0))
	if err != nil {
		return nil, fmt.Errorf("failed to get genesis block header: %w", err)
	}

	iter := wal.NewIterator(ctx, w.GethInstance.WALDir, 1, genesisHeader.Hash(), false)

	ops := []wal.Operation{}

	for block, err := range iter {
		if err != nil {
			return nil, fmt.Errorf("failed to read write-ahead log: %w", err)
		}

		for op, err := range block.OperationsIterator {
			if err != nil {
				return nil, fmt.Errorf("failed to read write-ahead log operation for block %d: %w", block.BlockInfo.Number, err)
			}
			ops = append(ops, op)
		}
	}

	return ops, nil

}
