package entityexpiration

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/keyset"
	"github.com/holiman/uint256"
)

func ClearEntitiesToExpireAtBlock(access StateAccess, blockNumber uint64) {
	blockNumberBig := uint256.NewInt(blockNumber)
	expiredEntityKey := crypto.Keccak256Hash(BlockExpirationSalt, blockNumberBig.Bytes())
	keyset.Clear(access, expiredEntityKey)
}
