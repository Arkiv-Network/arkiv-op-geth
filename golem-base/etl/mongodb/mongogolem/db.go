package mongogolem

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoGolem struct {
	db *mongo.Database
}

// New creates a new MongoGolem instance
func New(db *mongo.Database) *MongoGolem {
	return &MongoGolem{
		db: db,
	}
}

// Collections returns the MongoDB collections
func (m *MongoGolem) Collections() struct {
	ProcessingStatus *mongo.Collection
	Entities         *mongo.Collection
} {
	return struct {
		ProcessingStatus *mongo.Collection
		Entities         *mongo.Collection
	}{
		ProcessingStatus: m.db.Collection("processing_status"),
		Entities:         m.db.Collection("entities"),
	}
}

// EnsureIndexes creates all needed indexes for the collections
func (m *MongoGolem) EnsureIndexes(ctx context.Context) error {
	cols := m.Collections()

	// Create network index for processing_status
	networkIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "network", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := cols.ProcessingStatus.Indexes().CreateOne(ctx, networkIndex)
	if err != nil {
		return fmt.Errorf("failed to create network index for processing_status: %w", err)
	}

	// Create expiration index for entities
	expirationIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "expires_at", Value: 1}},
	}
	_, err = cols.Entities.Indexes().CreateOne(ctx, expirationIndex)
	if err != nil {
		return fmt.Errorf("failed to create expiration index for entities: %w", err)
	}

	// Create owner address index for entities
	ownerAddressIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "owner_address", Value: 1}},
	}
	_, err = cols.Entities.Indexes().CreateOne(ctx, ownerAddressIndex)
	if err != nil {
		return fmt.Errorf("failed to create owner address index for entities: %w", err)
	}

	// Create wildcard index for string annotations
	stringAnnotationsIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "stringAnnotations.$**", Value: 1}},
	}
	_, err = cols.Entities.Indexes().CreateOne(ctx, stringAnnotationsIndex)
	if err != nil {
		return fmt.Errorf("failed to create wildcard index for string annotations: %w", err)
	}

	// Create wildcard index for numeric annotations
	numericAnnotationsIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "numericAnnotations.$**", Value: 1}},
	}
	_, err = cols.Entities.Indexes().CreateOne(ctx, numericAnnotationsIndex)
	if err != nil {
		return fmt.Errorf("failed to create wildcard index for numeric annotations: %w", err)
	}

	return nil
}
