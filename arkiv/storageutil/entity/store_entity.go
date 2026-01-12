package entity

import (
	"fmt"
	"regexp"

	"github.com/ethereum/go-ethereum/arkiv/storageutil"
	"github.com/ethereum/go-ethereum/arkiv/storageutil/entity/entityexpiration"
	"github.com/ethereum/go-ethereum/common"
)

// This regex should not allow $ or 0x as the first characters, since we use that for
// special meta-annotations like $owner and for hashes and addresses.
const AnnotationIdentRegex string = `[\p{L}_][\p{L}\p{N}_]*`

var AnnotationIdentRegexCompiled *regexp.Regexp = regexp.MustCompile(fmt.Sprintf("^%s$", AnnotationIdentRegex))

type StateAccess = storageutil.StateAccess

func Store(
	access StateAccess,
	key common.Hash,
	sender common.Address,
	emd EntityMetaData,
	payload []byte,
) error {

	err := StoreEntityMetaData(access, key, emd)
	if err != nil {
		return fmt.Errorf("failed to store entity meta data: %w", err)
	}

	err = entityexpiration.AddToEntitiesToExpireAtBlock(access, emd.ExpiresAtBlock, key)
	if err != nil {
		return fmt.Errorf("failed to add entity to entities to expire: %w", err)
	}

	return nil
}
