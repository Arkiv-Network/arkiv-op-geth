package entity

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/stateblob"
)

func DeletePayload(access StateAccess, key common.Hash) {
	hash := crypto.Keccak256Hash(PayloadSalt, key[:])
	stateblob.DeleteBlob(access, hash)
}
