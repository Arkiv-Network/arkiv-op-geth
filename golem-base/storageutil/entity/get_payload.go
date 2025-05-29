package entity

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/stateblob"
)

func GetPayload(access StateAccess, key common.Hash) []byte {
	hash := crypto.Keccak256Hash(PayloadSalt, key[:])
	return stateblob.GetBlob(access, hash)
}
