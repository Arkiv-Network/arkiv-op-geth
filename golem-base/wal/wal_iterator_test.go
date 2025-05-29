package wal_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/golem-base/wal"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

func writeWal(
	dir string,
	blockInfo wal.BlockInfo,
	operations []wal.Operation,
) (err error) {

	f, err := os.Create(filepath.Join(dir, wal.BlockNumberToFilename(blockInfo.Number)+".temp"))
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)

	err = enc.Encode(blockInfo)
	if err != nil {
		return fmt.Errorf("failed to encode block info: %w", err)
	}

	for _, operation := range operations {
		err = enc.Encode(operation)
		if err != nil {
			return fmt.Errorf("failed to encode operation: %w", err)
		}
	}

	err = f.Close()
	if err != nil {
		return err
	}

	err = os.Rename(filepath.Join(dir, wal.BlockNumberToFilename(blockInfo.Number)+".temp"), filepath.Join(dir, wal.BlockNumberToFilename(blockInfo.Number)))
	if err != nil {
		return err
	}

	return nil
}

func TestWalIterator(t *testing.T) {

	t.Run("should iterate over one block", func(t *testing.T) {

		log.SetDefault(log.NewLogger(slog.NewTextHandler(os.Stdout, nil)))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		td := t.TempDir()

		err := writeWal(td,
			wal.BlockInfo{
				Number:     0,
				Hash:       common.HexToHash("0x1"),
				ParentHash: common.Hash{},
			},
			[]wal.Operation{
				{
					Create: &wal.Create{
						EntityKey:      common.HexToHash("0x1"),
						Payload:        []byte{1, 2, 3},
						ExpiresAtBlock: 100,
					},
				},
			},
		)
		require.NoError(t, err)

		for block, err := range wal.NewIterator(ctx, td, 0, common.Hash{}, false) {
			require.Equal(t, block.BlockInfo.Number, uint64(0))
			for operation, err := range block.OperationsIterator {
				require.NoError(t, err)
				require.Equal(t, operation.Create.EntityKey, common.HexToHash("0x1"))
				require.Equal(t, operation.Create.Payload, []byte{1, 2, 3})
				require.Equal(t, operation.Create.ExpiresAtBlock, uint64(100))
			}

			require.NoError(t, err)
			cancel()
		}
	})

	t.Run("should stop waiting for new blocks when context is cancelled", func(t *testing.T) {

		log.SetDefault(log.NewLogger(slog.NewTextHandler(os.Stdout, nil)))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		td := t.TempDir()

		err := writeWal(td,
			wal.BlockInfo{
				Number:     0,
				Hash:       common.HexToHash("0x1"),
				ParentHash: common.Hash{},
			},
			[]wal.Operation{
				{
					Create: &wal.Create{
						EntityKey:      common.HexToHash("0x1"),
						Payload:        []byte{1, 2, 3},
						ExpiresAtBlock: 100,
					},
				},
			},
		)
		require.NoError(t, err)

		go func() {
			time.Sleep(time.Second)
			cancel()
		}()

		for block, err := range wal.NewIterator(ctx, td, 0, common.Hash{}, true) {
			require.NoError(t, err)
			for range block.OperationsIterator {
			}

		}
	})

}
