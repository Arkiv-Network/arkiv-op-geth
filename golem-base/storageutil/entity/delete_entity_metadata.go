package entity

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/golem-base/storageutil"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/stateblob"
)

func DeleteEntityMetadata(access storageutil.StateAccess, key common.Hash) {

	hash := crypto.Keccak256Hash(EntityMetaDataSalt, key[:])
	stateblob.DeleteBlob(access, hash)
}
