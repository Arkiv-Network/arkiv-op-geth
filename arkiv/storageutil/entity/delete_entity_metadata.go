package entity

import (
	"github.com/ethereum/go-ethereum/arkiv/address"
	"github.com/ethereum/go-ethereum/arkiv/storageutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func DeleteEntityMetadata(access storageutil.StateAccess, key common.Hash) {

	hash := crypto.Keccak256Hash(EntityMetaDataSalt, key[:])
	access.SetState(address.ArkivProcessorAddress, hash, common.Hash{})
}
