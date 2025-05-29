package annotationindex

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var StringAnnotationIndexSalt = []byte("golemBaseStringAnnotation")
var AnnotationSeparator = []byte("|")

func StringAnnotationIndexKey(key, value string) common.Hash {
	return crypto.Keccak256Hash(StringAnnotationIndexSalt, []byte(key), AnnotationSeparator, []byte(value))
}
