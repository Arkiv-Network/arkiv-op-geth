package mongogolem

import (
	"time"
)

// ProcessingStatus tracks the last processed block
type ProcessingStatus struct {
	Network                  string    `bson:"network"`
	LastProcessedBlockNumber int64     `bson:"last_processed_block_number"`
	LastProcessedBlockHash   string    `bson:"last_processed_block_hash"`
	UpdatedAt                time.Time `bson:"updated_at"`
}

// Entity represents a stored entity with embedded annotations
type Entity struct {
	Key                string            `bson:"_id"`
	ExpiresAt          int64             `bson:"expires_at"`
	Payload            []byte            `bson:"content"`
	PayloadAsJSON      interface{}       `bson:"content_json,omitempty"`
	StringAnnotations  map[string]string `bson:"stringAnnotations,omitempty"`
	NumericAnnotations map[string]int64  `bson:"numericAnnotations,omitempty"`
	CreatedAt          time.Time         `bson:"created_at"`
	UpdatedAt          time.Time         `bson:"updated_at"`
	OwnerAddress       string            `bson:"owner_address"`
}

// Annotation represents a key-value pair
type Annotation struct {
	Key   string
	Value interface{}
}

// StringAnnotation represents a string annotation for an entity
type StringAnnotation struct {
	Key   string
	Value string
}

// NumericAnnotation represents a numeric annotation for an entity
type NumericAnnotation struct {
	Key   string
	Value int64
}
