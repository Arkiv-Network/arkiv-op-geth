package entity

import (
	"github.com/ethereum/go-ethereum/arkiv/address"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func StoreEntityMetaData(access StateAccess, key common.Hash, emd EntityMetaData) error {

	access.SetState(
		address.ArkivProcessorAddress,
		crypto.Keccak256Hash(EntityMetaDataSalt, key[:]),
		emd.Marshal(),
	)

	return nil

}
