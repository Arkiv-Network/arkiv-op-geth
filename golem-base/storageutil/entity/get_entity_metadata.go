package entity

import (
	"fmt"

	"github.com/ethereum/go-ethereum/arkiv/compression"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/entity/allentities"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/stateblob"
	"github.com/ethereum/go-ethereum/rlp"
)

var EntityMetaDataSalt = []byte("arkivEntityMetaData")

func GetEntityMetaData(access StateAccess, key common.Hash) (*EntityMetaData, error) {

	if !allentities.Contains(access, key) {
		return nil, fmt.Errorf("entity %s not found", key.Hex())
	}

	hash := crypto.Keccak256Hash(EntityMetaDataSalt, key[:])
	d := stateblob.GetBlob(access, hash)

	decoded, err := compression.BrotliDecompress(d)
	if err != nil {
		return nil, fmt.Errorf("failed to decode compressed entity meta data: %w", err)
	}

	emd := EntityMetaData{}
	err = rlp.DecodeBytes(decoded, &emd)
	if err != nil {
		return nil, err
	}

	return &emd, nil

}
