package golemtype

import (
	"github.com/ethereum/go-ethereum/common"
)

type SearchResult struct {
	Key   common.Hash `json:"key"`
	Value []byte      `json:"value"`
}
