package wal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type BlockOperationsIterator func(yield func(operation Operation, err error) bool)

func NewBlockOperationsIterator(ctx context.Context, path string) (BlockInfo, BlockOperationsIterator, error) {
	f, err := os.Open(path)
	if err != nil {
		return BlockInfo{}, nil, err
	}

	dec := json.NewDecoder(f)
	bi := BlockInfo{}
	err = dec.Decode(&bi)
	if err != nil {
		f.Close()
		return BlockInfo{}, nil, fmt.Errorf("failed to decode block info: %w", err)
	}

	return bi,
		BlockOperationsIterator(
			func(yield func(operation Operation, err error) bool) {
				defer f.Close()
				for ctx.Err() == nil {
					operation := Operation{}
					err = dec.Decode(&operation)
					if err == io.EOF {
						return
					}
					if err != nil {
						if !yield(Operation{}, err) {
							return
						}
					}
					if !yield(operation, nil) {
						return
					}
				}
				yield(Operation{}, ctx.Err())
			},
		),
		nil

}
