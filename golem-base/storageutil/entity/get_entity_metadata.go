package entity

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/entity/allentities"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/stateblob"
	"github.com/ethereum/go-ethereum/rlp"
)

var EntityMetaDataSalt = []byte("golemBaseEntityMetaData")

func GetEntityMetaData(access StateAccess, key common.Hash) (*EntityMetaData, error) {

	if !allentities.Contains(access, key) {
		return nil, fmt.Errorf("entity %s not found", key.Hex())
	}

	hash := crypto.Keccak256Hash(EntityMetaDataSalt, key[:])
	d := stateblob.GetBlob(access, hash)

	emd := EntityMetaData{}
	err := rlp.DecodeBytes(d, &emd)
	if err != nil {
		return nil, err
	}

	return &emd, nil

}
