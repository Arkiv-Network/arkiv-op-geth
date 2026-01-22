package entity

import (
	"fmt"

	"github.com/ethereum/go-ethereum/arkiv/address"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var EntityMetaDataSalt = []byte("arkivEntityMetaData")

func GetEntityMetaData(access StateAccess, key common.Hash) (*EntityMetaData, error) {
	value := access.GetState(address.ArkivProcessorAddress, crypto.Keccak256Hash(EntityMetaDataSalt, key[:]))

	if value == (common.Hash{}) {
		return nil, fmt.Errorf("failed to retrieve entity metadata for key %s", key.Hex())
	}

	emd := &EntityMetaData{}
	emd.Unmarshal(value)

	return emd, nil
}
