package entity

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/golem-base/storageutil"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/entity/allentities"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/entity/annotationindex"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/entity/entitiesofowner"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/entity/entityexpiration"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/keyset"
)

type StateAccess = storageutil.StateAccess

func Store(
	access StateAccess,
	key common.Hash,
	sender common.Address,
	emd EntityMetaData,
	payload []byte,
) error {

	err := allentities.AddEntity(access, key)
	if err != nil {
		return fmt.Errorf("failed to add entity to all entities: %w", err)
	}

	err = entitiesofowner.AddEntity(access, sender, key)
	if err != nil {
		return fmt.Errorf("failed to add entity to owner entities: %w", err)
	}

	err = StoreEntityMetaData(access, key, emd)
	if err != nil {
		return fmt.Errorf("failed to store entity meta data: %w", err)
	}

	err = entityexpiration.AddToEntitiesToExpireAtBlock(access, emd.ExpiresAtBlock, key)
	if err != nil {
		return fmt.Errorf("failed to add entity to entities to expire: %w", err)
	}

	for _, stringAnnotation := range emd.StringAnnotations {
		err = keyset.AddValue(
			access,
			annotationindex.StringAnnotationIndexKey(stringAnnotation.Key, stringAnnotation.Value),
			key,
		)
		if err != nil {
			return fmt.Errorf("failed to append to key list: %w", err)
		}
	}

	for _, numericAnnotation := range emd.NumericAnnotations {
		err = keyset.AddValue(
			access,
			annotationindex.NumericAnnotationIndexKey(numericAnnotation.Key, numericAnnotation.Value),
			key,
		)
		if err != nil {
			return fmt.Errorf("failed to append to key list: %w", err)
		}
	}

	StorePayload(access, key, payload)

	return nil
}
