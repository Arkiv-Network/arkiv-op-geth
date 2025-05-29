package entityexpiration

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/keyset"
	"github.com/holiman/uint256"
)

func IteratorOfEntitiesToExpireAtBlock(access StateAccess, blockNumber uint64) func(yield func(value common.Hash) bool) {
	blockNumberBig := uint256.NewInt(blockNumber)

	expiredEntityKey := crypto.Keccak256Hash(BlockExpirationSalt, blockNumberBig.Bytes())

	return keyset.Iterate(access, expiredEntityKey)

}
