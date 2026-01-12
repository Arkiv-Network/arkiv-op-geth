package entity

import (
	"fmt"

	"github.com/ethereum/go-ethereum/arkiv/storageutil"
	"github.com/ethereum/go-ethereum/arkiv/storageutil/entity/entityexpiration"
	"github.com/ethereum/go-ethereum/common"
)

func ExtendBTL(
	access storageutil.StateAccess,
	entityKey common.Hash,
	numberOfBlocks uint64) (uint64, common.Address, error) {

	entity, err := GetEntityMetaData(access, entityKey)
	if err != nil {
		return 0, common.Address{}, err
	}

	err = entityexpiration.RemoveFromEntitiesToExpire(access, entity.ExpiresAtBlock, entityKey)
	if err != nil {
		return 0, common.Address{}, fmt.Errorf("failed to remove from entities to expire at block %d: %w", entity.ExpiresAtBlock, err)
	}

	oldExpiresAtBlock := entity.ExpiresAtBlock

	entity.ExpiresAtBlock += numberOfBlocks

	err = entityexpiration.AddToEntitiesToExpireAtBlock(access, entity.ExpiresAtBlock, entityKey)
	if err != nil {
		return 0, common.Address{}, fmt.Errorf("failed to add to entities to expire at block %d: %w", entity.ExpiresAtBlock, err)
	}

	err = StoreEntityMetaData(access, entityKey, *entity)
	if err != nil {
		return 0, common.Address{}, fmt.Errorf("failed to store entity meta data: %w", err)
	}

	return oldExpiresAtBlock, entity.Owner, nil

}
