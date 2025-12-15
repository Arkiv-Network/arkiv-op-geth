package entity

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/golem-base/address"
	"github.com/ethereum/go-ethereum/golem-base/storageutil"
)

func DeleteEntityMetadata(access storageutil.StateAccess, key common.Hash) {

	hash := crypto.Keccak256Hash(EntityMetaDataSalt, key[:])
	access.SetState(address.ArkivProcessorAddress, hash, common.Hash{})
}
