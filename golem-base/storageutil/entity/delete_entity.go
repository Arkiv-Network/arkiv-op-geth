package entity

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/entity/allentities"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/entity/annotationindex"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/entity/entitiesofowner"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/entity/entityexpiration"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/keyset"
)

func Delete(access StateAccess, toDelete common.Hash) error {

	if !allentities.Contains(access, toDelete) {
		return fmt.Errorf("entity %s does not exist", toDelete)
	}

	md, err := GetEntityMetaData(access, toDelete)
	if err != nil {
		return fmt.Errorf("failed to get entity meta data: %w", err)
	}

	err = allentities.RemoveEntity(access, toDelete)
	if err != nil {
		return fmt.Errorf("failed to remove entity from all entities: %w", err)
	}

	for _, stringAnnotation := range md.StringAnnotations {
		setKey := annotationindex.StringAnnotationIndexKey(stringAnnotation.Key, stringAnnotation.Value)
		err := keyset.RemoveValue(
			access,
			setKey,
			toDelete,
		)
		if err != nil {
			return fmt.Errorf("failed to remove key %s from the string annotation list: %w", toDelete, err)
		}

	}

	for _, numericAnnotation := range md.NumericAnnotations {
		setKeys := annotationindex.NumericAnnotationIndexKey(numericAnnotation.Key, numericAnnotation.Value)
		err := keyset.RemoveValue(
			access,
			setKeys,
			toDelete,
		)
		if err != nil {
			return fmt.Errorf("failed to remove key %s from the numeric annotation list: %w", toDelete, err)
		}
	}

	err = entityexpiration.RemoveFromEntitiesToExpire(access, md.ExpiresAtBlock, toDelete)
	if err != nil {
		return fmt.Errorf("failed to remove entity from entities to expire: %w", err)
	}

	err = entitiesofowner.RemoveEntity(access, md.Owner, toDelete)
	if err != nil {
		return fmt.Errorf("failed to remove entity from owner entities: %w", err)
	}

	DeletePayload(access, toDelete)

	return nil
}
