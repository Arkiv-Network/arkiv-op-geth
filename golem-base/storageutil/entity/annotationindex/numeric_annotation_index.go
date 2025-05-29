package annotationindex

import (
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var NumericAnnotationIndexSalt = []byte("golemBaseNumericAnnotation")

func NumericAnnotationIndexKey(key string, value uint64) common.Hash {
	return crypto.Keccak256Hash(NumericAnnotationIndexSalt, []byte(key), AnnotationSeparator, binary.BigEndian.AppendUint64(nil, value))
}
