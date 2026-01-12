package entity

import (
	"github.com/ethereum/go-ethereum/arkiv/address"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var EntityMetaDataSalt = []byte("arkivEntityMetaData")

func GetEntityMetaData(access StateAccess, key common.Hash) (*EntityMetaData, error) {

	value := access.GetState(address.ArkivProcessorAddress, crypto.Keccak256Hash(EntityMetaDataSalt, key[:]))

	emd := &EntityMetaData{}
	emd.Unmarshal(value)

	return emd, nil

}
