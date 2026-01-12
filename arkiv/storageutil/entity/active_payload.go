package entity

import (
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
)

// EntityMetaData represents information about an entity that is currently active in the storage layer.
// This is what is stored in the state.
// It contains a BTL (number of blocks) and a list of annotations.
// The Key of the entity is derived from the payload content and the transaction hash where the entity was created.
type EntityMetaData struct {
	Owner          common.Address `json:"owner"`
	ExpiresAtBlock uint64         `json:"expiresAtBlock"`
}

func (emd *EntityMetaData) Marshal() common.Hash {
	bytes := [32]byte{}
	copy(bytes[:], emd.Owner[:])
	binary.BigEndian.PutUint64(bytes[24:], emd.ExpiresAtBlock)
	return bytes
}

func (emd *EntityMetaData) Unmarshal(hash common.Hash) {
	emd.Owner = common.BytesToAddress(hash[:20])
	emd.ExpiresAtBlock = binary.BigEndian.Uint64(hash[24:])
}
