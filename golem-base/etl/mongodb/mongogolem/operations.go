package mongogolem

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// HasProcessingStatus checks if a processing status exists for the given network
func (m *MongoGolem) HasProcessingStatus(ctx context.Context, network string) (bool, error) {
	cols := m.Collections()

	count, err := cols.ProcessingStatus.CountDocuments(ctx, bson.M{"network": network})
	if err != nil {
		return false, fmt.Errorf("failed to count processing status: %w", err)
	}

	return count > 0, nil
}

// GetProcessingStatus retrieves the processing status for the given network
func (m *MongoGolem) GetProcessingStatus(ctx context.Context, network string) (ProcessingStatus, error) {
	cols := m.Collections()

	var status ProcessingStatus
	err := cols.ProcessingStatus.FindOne(ctx, bson.M{"network": network}).Decode(&status)
	if err != nil {
		return ProcessingStatus{}, fmt.Errorf("failed to get processing status: %w", err)
	}

	return status, nil
}

// InsertProcessingStatus inserts a new processing status
func (m *MongoGolem) InsertProcessingStatus(ctx context.Context, params ProcessingStatus) error {
	cols := m.Collections()

	params.UpdatedAt = time.Now()

	_, err := cols.ProcessingStatus.InsertOne(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to insert processing status: %w", err)
	}

	return nil
}

// UpdateProcessingStatus updates the processing status
func (m *MongoGolem) UpdateProcessingStatus(ctx context.Context, params ProcessingStatus) error {
	cols := m.Collections()

	params.UpdatedAt = time.Now()

	_, err := cols.ProcessingStatus.UpdateOne(
		ctx,
		bson.M{"network": params.Network},
		bson.M{"$set": params},
	)
	if err != nil {
		return fmt.Errorf("failed to update processing status: %w", err)
	}

	return nil
}

// GetEntity retrieves an entity by key
func (m *MongoGolem) GetEntity(ctx context.Context, key string) (Entity, error) {
	cols := m.Collections()

	var entity Entity
	err := cols.Entities.FindOne(ctx, bson.M{"_id": key}).Decode(&entity)
	if err != nil {
		return Entity{}, fmt.Errorf("failed to get entity: %w", err)
	}

	return entity, nil
}

// InsertEntity inserts a new entity
func (m *MongoGolem) InsertEntity(ctx context.Context, entity Entity) error {
	cols := m.Collections()

	now := time.Now()
	entity.CreatedAt = now
	entity.UpdatedAt = now

	// Initialize annotations maps if they're nil
	if entity.StringAnnotations == nil {
		entity.StringAnnotations = make(map[string]string)
	}
	if entity.NumericAnnotations == nil {
		entity.NumericAnnotations = make(map[string]int64)
	}

	// Try to deserialize the payload to JSON if it's not empty
	if len(entity.Payload) > 0 {
		var jsonData interface{}
		if err := json.Unmarshal(entity.Payload, &jsonData); err == nil {
			entity.PayloadAsJSON = jsonData
		}
	}

	_, err := cols.Entities.InsertOne(ctx, entity)
	if err != nil {
		return fmt.Errorf("failed to insert entity: %w", err)
	}

	return nil
}

// UpdateEntity updates an existing entity
func (m *MongoGolem) UpdateEntity(ctx context.Context, entity Entity) error {
	cols := m.Collections()

	entity.UpdatedAt = time.Now()

	// Initialize annotations maps if they're nil
	if entity.StringAnnotations == nil {
		entity.StringAnnotations = make(map[string]string)
	}
	if entity.NumericAnnotations == nil {
		entity.NumericAnnotations = make(map[string]int64)
	}

	// Try to deserialize the payload to JSON if it's not empty
	if len(entity.Payload) > 0 {
		var jsonData interface{}
		if err := json.Unmarshal(entity.Payload, &jsonData); err == nil {
			entity.PayloadAsJSON = jsonData
		}
	}

	_, err := cols.Entities.ReplaceOne(
		ctx,
		bson.M{"_id": entity.Key},
		entity,
	)
	if err != nil {
		return fmt.Errorf("failed to update entity: %w", err)
	}

	return nil
}

// DeleteEntity deletes an entity by key
func (m *MongoGolem) DeleteEntity(ctx context.Context, key string) error {
	cols := m.Collections()

	_, err := cols.Entities.DeleteOne(ctx, bson.M{"_id": key})
	if err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}

	return nil
}

// AddStringAnnotation adds a string annotation to an entity
func (m *MongoGolem) AddStringAnnotation(ctx context.Context, entityKey string, annotation StringAnnotation) error {
	cols := m.Collections()

	updateField := fmt.Sprintf("stringAnnotations.%s", annotation.Key)
	_, err := cols.Entities.UpdateOne(
		ctx,
		bson.M{"_id": entityKey},
		bson.M{
			"$set": bson.M{
				updateField:  annotation.Value,
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to add string annotation: %w", err)
	}

	return nil
}

// AddNumericAnnotation adds a numeric annotation to an entity
func (m *MongoGolem) AddNumericAnnotation(ctx context.Context, entityKey string, annotation NumericAnnotation) error {
	cols := m.Collections()

	updateField := fmt.Sprintf("numericAnnotations.%s", annotation.Key)
	_, err := cols.Entities.UpdateOne(
		ctx,
		bson.M{"_id": entityKey},
		bson.M{
			"$set": bson.M{
				updateField:  annotation.Value,
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to add numeric annotation: %w", err)
	}

	return nil
}

// AddStringAnnotations adds multiple string annotations to an entity
func (m *MongoGolem) AddStringAnnotations(ctx context.Context, entityKey string, annotations []StringAnnotation) error {
	if len(annotations) == 0 {
		return nil
	}

	cols := m.Collections()
	update := bson.M{"updated_at": time.Now()}

	for _, annotation := range annotations {
		updateField := fmt.Sprintf("stringAnnotations.%s", annotation.Key)
		update[updateField] = annotation.Value
	}

	_, err := cols.Entities.UpdateOne(
		ctx,
		bson.M{"_id": entityKey},
		bson.M{"$set": update},
	)
	if err != nil {
		return fmt.Errorf("failed to add string annotations: %w", err)
	}

	return nil
}

// AddNumericAnnotations adds multiple numeric annotations to an entity
func (m *MongoGolem) AddNumericAnnotations(ctx context.Context, entityKey string, annotations []NumericAnnotation) error {
	if len(annotations) == 0 {
		return nil
	}

	cols := m.Collections()
	update := bson.M{"updated_at": time.Now()}

	for _, annotation := range annotations {
		updateField := fmt.Sprintf("numericAnnotations.%s", annotation.Key)
		update[updateField] = annotation.Value
	}

	_, err := cols.Entities.UpdateOne(
		ctx,
		bson.M{"_id": entityKey},
		bson.M{"$set": update},
	)
	if err != nil {
		return fmt.Errorf("failed to add numeric annotations: %w", err)
	}

	return nil
}

// RemoveStringAnnotation removes a string annotation from an entity
func (m *MongoGolem) RemoveStringAnnotation(ctx context.Context, entityKey string, annotationKey string) error {
	cols := m.Collections()

	updateField := fmt.Sprintf("stringAnnotations.%s", annotationKey)
	_, err := cols.Entities.UpdateOne(
		ctx,
		bson.M{"_id": entityKey},
		bson.M{
			"$unset": bson.M{updateField: ""},
			"$set":   bson.M{"updated_at": time.Now()},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to remove string annotation: %w", err)
	}

	return nil
}

// RemoveNumericAnnotation removes a numeric annotation from an entity
func (m *MongoGolem) RemoveNumericAnnotation(ctx context.Context, entityKey string, annotationKey string) error {
	cols := m.Collections()

	updateField := fmt.Sprintf("numericAnnotations.%s", annotationKey)
	_, err := cols.Entities.UpdateOne(
		ctx,
		bson.M{"_id": entityKey},
		bson.M{
			"$unset": bson.M{updateField: ""},
			"$set":   bson.M{"updated_at": time.Now()},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to remove numeric annotation: %w", err)
	}

	return nil
}

// ClearStringAnnotations removes all string annotations from an entity
func (m *MongoGolem) ClearStringAnnotations(ctx context.Context, entityKey string) error {
	cols := m.Collections()

	_, err := cols.Entities.UpdateOne(
		ctx,
		bson.M{"_id": entityKey},
		bson.M{
			"$set": bson.M{
				"stringAnnotations": make(map[string]string),
				"updated_at":        time.Now(),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to clear string annotations: %w", err)
	}

	return nil
}

// ClearNumericAnnotations removes all numeric annotations from an entity
func (m *MongoGolem) ClearNumericAnnotations(ctx context.Context, entityKey string) error {
	cols := m.Collections()

	_, err := cols.Entities.UpdateOne(
		ctx,
		bson.M{"_id": entityKey},
		bson.M{
			"$set": bson.M{
				"numericAnnotations": make(map[string]int64),
				"updated_at":         time.Now(),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to clear numeric annotations: %w", err)
	}

	return nil
}
