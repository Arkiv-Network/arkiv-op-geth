package entity

import (
	"fmt"

	"github.com/ethereum/go-ethereum/arkiv/storageutil/entity/entityexpiration"
	"github.com/ethereum/go-ethereum/common"
)

func Delete(access StateAccess, toDelete common.Hash) (common.Address, error) {

	md, err := GetEntityMetaData(access, toDelete)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to get entity meta data: %w", err)
	}

	err = entityexpiration.RemoveFromEntitiesToExpire(access, md.ExpiresAtBlock, toDelete)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to remove entity from entities to expire: %w", err)
	}

	DeleteEntityMetadata(access, toDelete)

	return md.Owner, nil
}
