package sqlstore

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/arkiv/compression"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/golem-base/arkivtype"
	"github.com/ethereum/go-ethereum/golem-base/query"
	"github.com/ethereum/go-ethereum/golem-base/sqlstore/sqlitegolem"
	"github.com/ethereum/go-ethereum/golem-base/storageutil/entity"
	"github.com/ethereum/go-ethereum/log"
	_ "github.com/mattn/go-sqlite3"
)

const entitiesSchemaVersion = uint64(6)

type BlockWal struct {
	BlockInfo  BlockInfo
	Operations []Operation
}
type BlockInfo struct {
	Number     uint64      `json:"number,string"`
	Hash       common.Hash `json:"hash"`
	ParentHash common.Hash `json:"parentHash"`
}

type Operation struct {
	Create      *Create      `json:"create,omitempty"`
	Update      *Update      `json:"update,omitempty"`
	ChangeOwner *ChangeOwner `json:"changeOwner,omitempty"`
	Delete      *Delete      `json:"delete,omitempty"`
	Extend      *ExtendBTL   `json:"extend,omitempty"`
}

type Create struct {
	EntityKey          common.Hash                `json:"entityKey"`
	ExpiresAtBlock     uint64                     `json:"expiresAtBlock"`
	Payload            []byte                     `json:"payload"`
	ContentType        string                     `json:"contentType"`
	StringAnnotations  []entity.StringAnnotation  `json:"stringAnnotations"`
	NumericAnnotations []entity.NumericAnnotation `json:"numericAnnotations"`
	Owner              common.Address             `json:"owner"`
	TransactionIndex   uint64                     `json:"txIndex"`
	OperationIndex     uint64                     `json:"opIndex"`
}

type Update struct {
	EntityKey          common.Hash                `json:"entityKey"`
	ExpiresAtBlock     uint64                     `json:"expiresAtBlock"`
	Payload            []byte                     `json:"payload"`
	ContentType        string                     `json:"contentType"`
	StringAnnotations  []entity.StringAnnotation  `json:"stringAnnotations"`
	NumericAnnotations []entity.NumericAnnotation `json:"numericAnnotations"`
	TransactionIndex   uint64                     `json:"txIndex"`
	OperationIndex     uint64                     `json:"opIndex"`
}

type ChangeOwner struct {
	EntityKey        common.Hash    `json:"entityKey"`
	Owner            common.Address `json:"owner"`
	TransactionIndex uint64         `json:"txIndex"`
	OperationIndex   uint64         `json:"opIndex"`
}

type ExtendBTL struct {
	EntityKey        common.Hash `json:"entityKey"`
	OldExpiresAt     uint64      `json:"oldExpiresAt"`
	NewExpiresAt     uint64      `json:"newExpiresAt"`
	TransactionIndex uint64      `json:"txIndex"`
	OperationIndex   uint64      `json:"opIndex"`
}

type Delete struct {
	EntityKey        common.Hash `json:"entityKey"`
	TransactionIndex uint64      `json:"txIndex"`
	OperationIndex   uint64      `json:"opIndex"`
}

// SQLStore encapsulates the SQLite SQLStore functionality
type SQLStore struct {
	writeDB             *sql.DB
	readDB              *sql.DB
	historicBlocksCount uint64
	databaseDisabled    bool
}

func getSequence(createdAtBlock uint64, transactionIndexInBlock uint64, operationIndexInTransaction uint64) uint64 {
	return createdAtBlock<<32 | transactionIndexInBlock<<16 | operationIndexInTransaction
}

// configureDBPool configures the database connection pool.
// Keep constant set of open connections and never close them.
// This allows us to avoid `database is locked` error, when new connection is opened.
func configureDBPool(db *sql.DB, numThreads int) {
	db.SetMaxOpenConns(numThreads)
	db.SetMaxIdleConns(numThreads)
	db.SetConnMaxLifetime(0)
	db.SetConnMaxIdleTime(0)
}

// logDBStats logs database statistics for both read and write databases
func (e *SQLStore) logDBStats(operation string, phase string) {
	readStats := e.readDB.Stats()
	writeStats := e.writeDB.Stats()

	log.Info("database stats",
		"phase", phase,
		"operation", operation,
		"readDB_openConnections", readStats.OpenConnections,
		"readDB_inUse", readStats.InUse,
		"readDB_idle", readStats.Idle,
		"readDB_waitCount", readStats.WaitCount,
		"readDB_waitDuration", readStats.WaitDuration,
		"writeDB_openConnections", writeStats.OpenConnections,
		"writeDB_inUse", writeStats.InUse,
		"writeDB_idle", writeStats.Idle,
		"writeDB_waitCount", writeStats.WaitCount,
		"writeDB_waitDuration", writeStats.WaitDuration,
	)
}

// NewStore creates a new ETL instance with database connection and schema setup
func NewStore(dbFile string, historicBlocksCount uint64, databaseDisabled bool) (*SQLStore, error) {
	log.Info("creating new SQLStore", "dbFile", dbFile, "historicBlocksCount", historicBlocksCount, "databaseDisabled", databaseDisabled)
	dir := filepath.Dir(dbFile)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=rwc&_busy_timeout=11000&_journal_mode=WAL&_auto_vacuum=incremental&_foreign_keys=true&_txlock=immediate&_cache_size=1000000000", dbFile))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	configureDBPool(db, 1)

	// Check if schema exists and apply if needed
	ctx := context.Background()

	// Check if schema is up to date
	readVersions := true
	entitiesVersion := uint64(0)

	var tableName string
	err = db.QueryRowContext(ctx, `
		SELECT name FROM sqlite_master
		WHERE type='table' AND name='schema_versions';
	`).Scan(&tableName)

	switch err {
	case sql.ErrNoRows:
		// In version 0, we didn't have the schema_versions table yet
		entitiesVersion = 0
		readVersions = false
		log.Warn("arkiv: no schema version info found, table missing")
	case nil:
		// The schema exists, we can read the versions from it
	default:
		// We got another error
		db.Close()
		return nil, fmt.Errorf("failed to check schema: %w", err)
	}

	if readVersions {
		err = db.QueryRowContext(
			ctx,
			`SELECT entities FROM schema_versions WHERE id = 1;`,
		).Scan(&entitiesVersion)

		switch err {
		case sql.ErrNoRows:
			entitiesVersion = 0
			log.Warn("arkiv: no schema version info found, table empty", "error", err)
		case nil:
			// We read the versions, all good
			log.Info("arkiv: schema versions read from database", "entities", entitiesVersion)
		default:
			db.Close()
			return nil, fmt.Errorf("failed to check schema: %w", err)
		}
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	if entitiesVersion != entitiesSchemaVersion {
		log.Warn(
			"arkiv: entities table has an outdated schema, dropping tables",
			"existingVersion", entitiesVersion,
			"requiredVersion", entitiesSchemaVersion,
		)
		_, err = tx.ExecContext(ctx, `DROP TABLE IF EXISTS string_annotations;`)
		if err != nil {
			tx.Rollback()
			db.Close()
			return nil, fmt.Errorf("failed to drop string_annotations table: %w", err)
		}
		_, err = tx.ExecContext(ctx, `DROP TABLE IF EXISTS numeric_annotations;`)
		if err != nil {
			tx.Rollback()
			db.Close()
			return nil, fmt.Errorf("failed to drop numeric_annotations table: %w", err)
		}
		_, err = tx.ExecContext(ctx, `DROP TABLE IF EXISTS entities;`)
		if err != nil {
			tx.Rollback()
			db.Close()
			return nil, fmt.Errorf("failed to drop entities table: %w", err)
		}
		_, err = tx.ExecContext(ctx, `DROP TABLE IF EXISTS processing_status;`)
		if err != nil {
			tx.Rollback()
			db.Close()
			return nil, fmt.Errorf("failed to drop processing_status table: %w", err)
		}
	}

	log.Info("arkiv: applying database schema")
	err = sqlitegolem.ApplySchemaTx(ctx, tx)
	if err != nil {
		tx.Rollback()
		db.Close()
		return nil, fmt.Errorf("failed to recreate schema: %w", err)
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT OR REPLACE INTO schema_versions (id, entities) VALUES (1, ?);`,
		entitiesSchemaVersion)
	if err != nil {
		tx.Rollback()
		db.Close()
		return nil, fmt.Errorf("failed to update schema versions: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		db.Close()
		return nil, fmt.Errorf("failed to recreate schema: %w", err)
	}

	readDB, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_query_only=true&_busy_timeout=11000&_journal_mode=WAL&_auto_vacuum=incremental&_foreign_keys=true&_txlock=deferred&_cache_size=1000000000", dbFile))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	configureDBPool(readDB, runtime.NumCPU())

	store := &SQLStore{
		writeDB:             db,
		readDB:              readDB,
		historicBlocksCount: historicBlocksCount,
		databaseDisabled:    databaseDisabled,
	}

	if !databaseDisabled {
		go store.collectGarbage()
	}

	log.Info("arkiv: database ready", "entitySchemaVersion", entitiesSchemaVersion)
	return store, nil
}

func (e *SQLStore) collectGarbage() {
	log.Info("started DB garbage collector")
	ctx := context.Background()
	for {
		time.Sleep(time.Minute)
		e.doCollectGarbage(ctx)
	}
}

func (e *SQLStore) doCollectGarbage(ctx context.Context) {
	readDB := sqlitegolem.New(e.readDB)

	blockNumber, err := readDB.GetLastProcessedBlockNumber(ctx)
	if err != nil {
		log.Error("failed to fetch current block number", "error", err)
		return
	}

	garbageCount, err := readDB.GetGarbageCount(ctx, blockNumber)
	if err != nil {
		log.Error("failed to fetch amount of garbage", "error", err)
		return
	}

	if garbageCount < 100 {
		log.Info("skipping garbage collection in the DB", "count", garbageCount)
		return
	}

	log.Info("collecting garbage in the DB", "count", garbageCount)

	e.logDBStats("collectGarbage", "start")
	tx, err := e.writeDB.BeginTx(ctx, nil)
	if err != nil {
		log.Error("failed to begin transaction", "error", err)
		return
	}

	txDB := sqlitegolem.New(tx)

	// Delete blocks that are older than the historicBlocksCount
	if e.historicBlocksCount > 0 && blockNumber > int64(e.historicBlocksCount) {
		deleteUntilBlock := blockNumber - int64(e.historicBlocksCount)

		err = errors.Join(
			txDB.DeleteStringAnnotationsUntilBlock(ctx, deleteUntilBlock),
			txDB.DeleteNumericAnnotationsUntilBlock(ctx, deleteUntilBlock),
			txDB.DeleteEntitiesUntilBlock(ctx, deleteUntilBlock),
		)
	}

	if err != nil {
		tx.Rollback()
		e.logDBStats("collectGarbage", "finish-rollback")
		log.Error("failed to collect garbage in DB", "error", err)
	} else {
		tx.Commit()
		e.logDBStats("collectGarbage", "finish-commit")
		log.Info("collected garbage in the DB")
	}
}

// Close closes the database connection
func (e *SQLStore) Close() error {
	return errors.Join(e.readDB.Close(), e.writeDB.Close())
}

// GetQueries returns a new sqlitegolem.Queries instance for autocommit operations
func (e *SQLStore) GetQueries() *sqlitegolem.Queries {
	return sqlitegolem.New(e.readDB)
}

func (e *SQLStore) GetProcessingStatus(ctx context.Context, networkID string) (*sqlitegolem.GetProcessingStatusRow, error) {
	result, err := e.GetQueries().GetProcessingStatus(ctx, networkID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &sqlitegolem.GetProcessingStatusRow{
				LastProcessedBlockNumber: 0,
				LastProcessedBlockHash:   "",
			}, nil
		}
		return nil, err
	}
	return &result, nil
}

// GetEntityCount retrieves the total number of entities in the database
func (e *SQLStore) GetEntityCount(ctx context.Context, block uint64) (uint64, error) {
	e.logDBStats("getEntityCount", "start")
	count, err := e.GetQueries().GetEntityCount(ctx, int64(block))
	if err != nil {
		e.logDBStats("getEntityCount", "finish-error")
		return 0, fmt.Errorf("failed to get entity count: %w", err)
	}

	e.logDBStats("getEntityCount", "finish-success")
	return uint64(count), nil
}

func (e *SQLStore) SnapSyncToBlock(
	ctx context.Context,
	networkID string,
	blockNumber uint64,
	blockHash common.Hash,
	entities iter.Seq2[
		*struct {
			Key      common.Hash
			Metadata entity.EntityMetaData
			Payload  []byte
		},
		error,
	],
) (err error) {
	if e.databaseDisabled {
		return nil
	}
	log.Info("snap syncing to block start", "blockNumber", blockNumber, "blockHash", blockHash.Hex())
	defer log.Info("snap syncing to block end", "blockNumber", blockNumber, "blockHash", blockHash.Hex())

	e.logDBStats("snapSyncToBlock", "start")
	tx, err := e.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback())
			e.logDBStats("snapSyncToBlock", "finish-rollback")
		}
	}()

	txDB := sqlitegolem.New(tx)

	// Ensure single network constraint for snap sync
	hasNetwork, err := txDB.HasProcessingStatus(ctx, networkID)
	if err != nil {
		return fmt.Errorf("failed to check if network exists: %w", err)
	}

	if !hasNetwork {
		// This is a new network, check if there are already other networks
		networkCount, err := txDB.CountNetworks(ctx)
		if err != nil {
			return fmt.Errorf("failed to count existing networks: %w", err)
		}

		if networkCount > 0 {
			return fmt.Errorf("cannot snap sync to network %s: database already contains %d network(s), only one network is allowed", networkID, networkCount)
		}

		// First network, need to insert initial processing status
		err = txDB.InsertProcessingStatus(ctx, sqlitegolem.InsertProcessingStatusParams{
			Network:                  networkID,
			LastProcessedBlockNumber: int64(blockNumber),
			LastProcessedBlockHash:   blockHash.Hex(),
		})
		if err != nil {
			return fmt.Errorf("failed to insert initial processing status: %w", err)
		}
	}

	// Clear all existing entities, annotations for a clean snap sync
	err = txDB.DeleteAllStringAnnotations(ctx)
	if err != nil {
		return fmt.Errorf("failed to clear string annotations: %w", err)
	}

	err = txDB.DeleteAllNumericAnnotations(ctx)
	if err != nil {
		return fmt.Errorf("failed to clear numeric annotations: %w", err)
	}

	err = txDB.DeleteAllEntities(ctx)
	if err != nil {
		return fmt.Errorf("failed to clear entities: %w", err)
	}

	// Insert all entities from the snapshot
	for entityToInsert, err := range entities {
		if err != nil {
			return fmt.Errorf("failed to get entity: %w", err)
		}

		// Insert the entity
		err = txDB.InsertEntity(ctx, sqlitegolem.InsertEntityParams{
			Key:                         strings.ToLower(entityToInsert.Key.Hex()),
			ExpiresAt:                   int64(entityToInsert.Metadata.ExpiresAtBlock),
			Payload:                     entityToInsert.Payload,
			ContentType:                 entityToInsert.Metadata.ContentType,
			OwnerAddress:                strings.ToLower(entityToInsert.Metadata.Owner.Hex()),
			CreatedAtBlock:              int64(entityToInsert.Metadata.CreatedAtBlock),
			LastModifiedAtBlock:         int64(entityToInsert.Metadata.LastModifiedAtBlock),
			TransactionIndexInBlock:     int64(entityToInsert.Metadata.TransactionIndex),
			OperationIndexInTransaction: int64(entityToInsert.Metadata.OperationIndex),
		})
		if err != nil {
			return fmt.Errorf("failed to insert entity %s: %w", entityToInsert.Key.Hex(), err)
		}

		// Insert string annotations
		strAnnotations := append(entityToInsert.Metadata.StringAnnotations,
			entity.StringAnnotation{
				Key:   arkivtype.KeyAttributeKey,
				Value: strings.ToLower(entityToInsert.Key.Hex()),
			},
			entity.StringAnnotation{
				Key:   arkivtype.OwnerAttributeKey,
				Value: strings.ToLower(entityToInsert.Metadata.Owner.Hex()),
			},
			entity.StringAnnotation{
				Key:   arkivtype.CreatorAttributeKey,
				Value: strings.ToLower(entityToInsert.Metadata.Creator.Hex()),
			},
		)
		for _, annotation := range strAnnotations {
			err = txDB.InsertStringAnnotation(ctx, sqlitegolem.InsertStringAnnotationParams{
				EntityKey:                         strings.ToLower(entityToInsert.Key.Hex()),
				EntityLastModifiedAtBlock:         int64(entityToInsert.Metadata.LastModifiedAtBlock),
				EntityTransactionIndexInBlock:     int64(entityToInsert.Metadata.TransactionIndex),
				EntityOperationIndexInTransaction: int64(entityToInsert.Metadata.OperationIndex),
				AnnotationKey:                     annotation.Key,
				Value:                             annotation.Value,
			})
			if err != nil {
				return fmt.Errorf("failed to insert string annotation for entity %s: %w", entityToInsert.Key.Hex(), err)
			}
		}

		// Insert numeric annotations
		numAnnotations := append(entityToInsert.Metadata.NumericAnnotations,
			entity.NumericAnnotation{
				Key:   arkivtype.ExpirationAttributeKey,
				Value: entityToInsert.Metadata.ExpiresAtBlock,
			},
			entity.NumericAnnotation{
				Key: arkivtype.SequenceAttributeKey,
				Value: getSequence(
					entityToInsert.Metadata.LastModifiedAtBlock,
					entityToInsert.Metadata.TransactionIndex,
					entityToInsert.Metadata.OperationIndex,
				),
			},
		)
		for _, annotation := range numAnnotations {
			err = txDB.InsertNumericAnnotation(ctx, sqlitegolem.InsertNumericAnnotationParams{
				EntityKey:                         strings.ToLower(entityToInsert.Key.Hex()),
				EntityLastModifiedAtBlock:         int64(entityToInsert.Metadata.LastModifiedAtBlock),
				EntityTransactionIndexInBlock:     int64(entityToInsert.Metadata.TransactionIndex),
				EntityOperationIndexInTransaction: int64(entityToInsert.Metadata.OperationIndex),
				AnnotationKey:                     annotation.Key,
				Value:                             int64(annotation.Value),
			})
			if err != nil {
				return fmt.Errorf("failed to insert numeric annotation for entity %s: %w", entityToInsert.Key.Hex(), err)
			}
		}
	}

	// Update processing status to the snap sync block
	err = txDB.UpdateProcessingStatus(ctx, sqlitegolem.UpdateProcessingStatusParams{
		Network:                  networkID,
		LastProcessedBlockNumber: int64(blockNumber),
		LastProcessedBlockHash:   blockHash.Hex(),
	})
	if err != nil {
		return fmt.Errorf("failed to update processing status: %w", err)
	}

	e.logDBStats("snapSyncToBlock", "before-commit")
	return tx.Commit()
}

// InsertBlock processes a single block from the WAL and inserts it into the database
func (e *SQLStore) InsertBlock(ctx context.Context, blockWal BlockWal, networkID string) (err error) {
	if e.databaseDisabled {
		return nil
	}
	log.Info("processing block", "block", blockWal.BlockInfo.Number)
	defer log.Info("processing block end", "block", blockWal.BlockInfo.Number)

	e.logDBStats("insertBlock", "start")
	tx, err := e.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback())
			e.logDBStats("insertBlock", "finish-rollback")
		}
	}()

	txDB := sqlitegolem.New(tx)

	// Ensure single network constraint: check if this would create a new network
	hasNetwork, err := txDB.HasProcessingStatus(ctx, networkID)
	if err != nil {
		return fmt.Errorf("failed to check if network exists: %w", err)
	}

	if !hasNetwork {
		// This is a new network, check if there are already other networks
		networkCount, err := txDB.CountNetworks(ctx)
		if err != nil {
			return fmt.Errorf("failed to count existing networks: %w", err)
		}

		if networkCount > 0 {
			return fmt.Errorf("cannot add network %s: database already contains %d network(s), only one network is allowed", networkID, networkCount)
		}

		err = txDB.InsertProcessingStatus(ctx, sqlitegolem.InsertProcessingStatusParams{
			Network:                  networkID,
			LastProcessedBlockNumber: int64(blockWal.BlockInfo.Number - 1),
			LastProcessedBlockHash:   blockWal.BlockInfo.ParentHash.Hex(),
		})
		if err != nil {
			return fmt.Errorf("failed to insert initial processing status: %w", err)
		}
	}

	log.Info("hasNetwork", "hasNetwork", hasNetwork)

	// Check if parent block hash matches the expected value from processing status
	if blockWal.BlockInfo.Number > 1 { // Skip check for genesis block
		processingStatus, err := txDB.GetProcessingStatus(ctx, networkID)
		if err != nil {
			return fmt.Errorf("failed to get processing status: %w", err)
		}

		expectedParentHash := processingStatus.LastProcessedBlockHash
		actualParentHash := blockWal.BlockInfo.ParentHash.Hex()

		if expectedParentHash != actualParentHash {
			return fmt.Errorf("parent block hash mismatch: expected %s, got %s", expectedParentHash, actualParentHash)
		}

		// Verify block number sequence
		expectedBlockNumber := processingStatus.LastProcessedBlockNumber + 1
		if int64(blockWal.BlockInfo.Number) != expectedBlockNumber {
			return fmt.Errorf("block number sequence error: expected %d, got %d", expectedBlockNumber, blockWal.BlockInfo.Number)
		}
	}

	for _, op := range blockWal.Operations {

		switch {
		case op.Create != nil:
			log.Info("create", "entity", op.Create.EntityKey.Hex())
			err = txDB.InsertEntity(ctx, sqlitegolem.InsertEntityParams{
				Key:                         strings.ToLower(op.Create.EntityKey.Hex()),
				ExpiresAt:                   int64(op.Create.ExpiresAtBlock),
				Payload:                     op.Create.Payload,
				ContentType:                 op.Create.ContentType,
				OwnerAddress:                strings.ToLower(op.Create.Owner.Hex()),
				CreatedAtBlock:              int64(blockWal.BlockInfo.Number),
				LastModifiedAtBlock:         int64(blockWal.BlockInfo.Number),
				TransactionIndexInBlock:     int64(op.Create.TransactionIndex),
				OperationIndexInTransaction: int64(op.Create.OperationIndex),
			})
			if err != nil {
				return fmt.Errorf("failed to insert entity: %w", err)
			}

			numAnnotations := append(op.Create.NumericAnnotations,
				entity.NumericAnnotation{
					Key:   arkivtype.ExpirationAttributeKey,
					Value: op.Create.ExpiresAtBlock,
				},
				entity.NumericAnnotation{
					Key: arkivtype.SequenceAttributeKey,
					Value: getSequence(
						blockWal.BlockInfo.Number,
						op.Create.TransactionIndex,
						op.Create.OperationIndex,
					),
				},
			)
			for _, annotation := range numAnnotations {
				err = txDB.InsertNumericAnnotation(ctx, sqlitegolem.InsertNumericAnnotationParams{
					EntityKey:                         strings.ToLower(op.Create.EntityKey.Hex()),
					EntityLastModifiedAtBlock:         int64(blockWal.BlockInfo.Number),
					EntityTransactionIndexInBlock:     int64(op.Create.TransactionIndex),
					EntityOperationIndexInTransaction: int64(op.Create.OperationIndex),
					AnnotationKey:                     annotation.Key,
					Value:                             int64(annotation.Value),
				})
				if err != nil {
					return fmt.Errorf("failed to insert numeric annotation: %w", err)
				}
			}

			strAnnotations := append(op.Create.StringAnnotations,
				entity.StringAnnotation{
					Key:   arkivtype.KeyAttributeKey,
					Value: strings.ToLower(op.Create.EntityKey.Hex()),
				},
				entity.StringAnnotation{
					Key:   arkivtype.OwnerAttributeKey,
					Value: strings.ToLower(op.Create.Owner.Hex()),
				},
				entity.StringAnnotation{
					Key:   arkivtype.CreatorAttributeKey,
					Value: strings.ToLower(op.Create.Owner.Hex()),
				},
			)
			for _, annotation := range strAnnotations {
				err = txDB.InsertStringAnnotation(ctx, sqlitegolem.InsertStringAnnotationParams{
					EntityKey:                         strings.ToLower(op.Create.EntityKey.Hex()),
					EntityLastModifiedAtBlock:         int64(blockWal.BlockInfo.Number),
					EntityTransactionIndexInBlock:     int64(op.Create.TransactionIndex),
					EntityOperationIndexInTransaction: int64(op.Create.OperationIndex),
					AnnotationKey:                     annotation.Key,
					Value:                             annotation.Value,
				})
				if err != nil {
					return fmt.Errorf("failed to insert string annotation: %w", err)
				}
			}
		case op.Update != nil:
			existingEntity, err := txDB.GetEntity(ctx, sqlitegolem.GetEntityParams{
				Key:   strings.ToLower(op.Update.EntityKey.Hex()),
				Block: int64(blockWal.BlockInfo.Number - 1),
			})
			if err != nil {
				return fmt.Errorf("failed to get existing entity: %w", err)
			}

			txDB.InsertEntity(ctx, sqlitegolem.InsertEntityParams{
				Key:                         strings.ToLower(op.Update.EntityKey.Hex()),
				ExpiresAt:                   int64(op.Update.ExpiresAtBlock),
				Payload:                     op.Update.Payload,
				ContentType:                 op.Update.ContentType,
				OwnerAddress:                strings.ToLower(existingEntity.OwnerAddress),
				CreatedAtBlock:              existingEntity.CreatedAtBlock,
				LastModifiedAtBlock:         int64(blockWal.BlockInfo.Number),
				Deleted:                     false,
				TransactionIndexInBlock:     int64(op.Update.TransactionIndex),
				OperationIndexInTransaction: int64(op.Update.OperationIndex),
			})

			numAnnotations := append(op.Update.NumericAnnotations,
				entity.NumericAnnotation{
					Key:   arkivtype.ExpirationAttributeKey,
					Value: op.Update.ExpiresAtBlock,
				},
				entity.NumericAnnotation{
					Key: arkivtype.SequenceAttributeKey,
					Value: getSequence(
						blockWal.BlockInfo.Number,
						op.Update.TransactionIndex,
						op.Update.OperationIndex,
					),
				},
			)
			for _, annotation := range numAnnotations {
				err = txDB.InsertNumericAnnotation(ctx, sqlitegolem.InsertNumericAnnotationParams{
					EntityKey:                         strings.ToLower(op.Update.EntityKey.Hex()),
					EntityLastModifiedAtBlock:         int64(blockWal.BlockInfo.Number),
					EntityTransactionIndexInBlock:     int64(op.Update.TransactionIndex),
					EntityOperationIndexInTransaction: int64(op.Update.OperationIndex),
					AnnotationKey:                     annotation.Key,
					Value:                             int64(annotation.Value),
				})
				if err != nil {
					return fmt.Errorf("failed to insert numeric annotation: %w", err)
				}
			}

			strAnnotations := append(op.Update.StringAnnotations,
				entity.StringAnnotation{
					Key:   arkivtype.KeyAttributeKey,
					Value: strings.ToLower(op.Update.EntityKey.Hex()),
				},
				entity.StringAnnotation{
					Key:   arkivtype.OwnerAttributeKey,
					Value: strings.ToLower(existingEntity.OwnerAddress),
				},
				entity.StringAnnotation{
					Key:   arkivtype.CreatorAttributeKey,
					Value: strings.ToLower(existingEntity.CreatorAddress),
				},
			)
			for _, annotation := range strAnnotations {
				err = txDB.InsertStringAnnotation(ctx, sqlitegolem.InsertStringAnnotationParams{
					EntityKey:                         strings.ToLower(op.Update.EntityKey.Hex()),
					EntityLastModifiedAtBlock:         int64(blockWal.BlockInfo.Number),
					EntityTransactionIndexInBlock:     int64(op.Update.TransactionIndex),
					EntityOperationIndexInTransaction: int64(op.Update.OperationIndex),
					AnnotationKey:                     annotation.Key,
					Value:                             annotation.Value,
				})
				if err != nil {
					return fmt.Errorf("failed to insert string annotation: %w", err)
				}
			}

		case op.ChangeOwner != nil:
			changeOwnerParams := sqlitegolem.UpdateEntityOwnerParams{
				Key:                         strings.ToLower(op.ChangeOwner.EntityKey.Hex()),
				LastModifiedAtBlock:         int64(blockWal.BlockInfo.Number),
				TransactionIndexInBlock:     int64(op.ChangeOwner.TransactionIndex),
				OperationIndexInTransaction: int64(op.ChangeOwner.OperationIndex),
				OwnerAddress:                strings.ToLower(op.ChangeOwner.Owner.Hex()),
			}

			log.Info("change owner", "params", changeOwnerParams)

			// Fetch the existing annotations before we update the entity, so that we
			// can re-insert them with the new block number.
			numericAnnotations, err := txDB.GetNumericAnnotations(ctx, sqlitegolem.GetNumericAnnotationsParams{
				EntityKey: strings.ToLower(op.ChangeOwner.EntityKey.Hex()),
				Block:     int64(blockWal.BlockInfo.Number),
			})
			if err != nil {
				return fmt.Errorf("failed to fetch annotations: %w", err)
			}

			stringAnnotations, err := txDB.GetStringAnnotations(ctx, sqlitegolem.GetStringAnnotationsParams{
				EntityKey: strings.ToLower(op.ChangeOwner.EntityKey.Hex()),
				Block:     int64(blockWal.BlockInfo.Number),
			})
			if err != nil {
				return fmt.Errorf("failed to fetch annotations: %w", err)
			}

			// Update the entity with the new expiry time
			err = txDB.UpdateEntityOwner(ctx, changeOwnerParams)
			if err != nil {
				return fmt.Errorf("failed to change owner: %w", err)
			}

			for _, annotation := range numericAnnotations {
				value := uint64(annotation.Value)
				if annotation.AnnotationKey == arkivtype.SequenceAttributeKey {
					value = getSequence(
						blockWal.BlockInfo.Number,
						op.ChangeOwner.TransactionIndex,
						op.ChangeOwner.OperationIndex,
					)
				}
				err = txDB.InsertNumericAnnotation(ctx, sqlitegolem.InsertNumericAnnotationParams{
					EntityKey:                         strings.ToLower(op.ChangeOwner.EntityKey.Hex()),
					EntityLastModifiedAtBlock:         int64(blockWal.BlockInfo.Number),
					EntityTransactionIndexInBlock:     int64(op.ChangeOwner.TransactionIndex),
					EntityOperationIndexInTransaction: int64(op.ChangeOwner.OperationIndex),
					AnnotationKey:                     annotation.AnnotationKey,
					Value:                             int64(value),
				})
				if err != nil {
					return fmt.Errorf("failed to insert numeric annotation: %w", err)
				}
			}

			for _, annotation := range stringAnnotations {
				value := annotation.Value
				if annotation.AnnotationKey == arkivtype.OwnerAttributeKey {
					value = op.ChangeOwner.Owner.Hex()
				}
				err = txDB.InsertStringAnnotation(ctx, sqlitegolem.InsertStringAnnotationParams{
					EntityKey:                         strings.ToLower(op.ChangeOwner.EntityKey.Hex()),
					EntityLastModifiedAtBlock:         int64(blockWal.BlockInfo.Number),
					EntityTransactionIndexInBlock:     int64(op.ChangeOwner.TransactionIndex),
					EntityOperationIndexInTransaction: int64(op.ChangeOwner.OperationIndex),
					AnnotationKey:                     annotation.AnnotationKey,
					Value:                             value,
				})
				if err != nil {
					return fmt.Errorf("failed to insert string annotation: %w", err)
				}
			}

		case op.Delete != nil:
			params := sqlitegolem.DeleteEntityParams{
				Key:                         strings.ToLower(op.Delete.EntityKey.Hex()),
				LastModifiedAtBlock:         int64(blockWal.BlockInfo.Number),
				TransactionIndexInBlock:     int64(op.Delete.TransactionIndex),
				OperationIndexInTransaction: int64(op.Delete.OperationIndex),
			}

			log.Info("delete entity", "params", params)

			err = txDB.DeleteEntity(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to delete entity: %w", err)
			}

		case op.Extend != nil:
			extendParams := sqlitegolem.UpdateEntityExpiresAtParams{
				Key:                         strings.ToLower(op.Extend.EntityKey.Hex()),
				LastModifiedAtBlock:         int64(blockWal.BlockInfo.Number),
				TransactionIndexInBlock:     int64(op.Extend.TransactionIndex),
				OperationIndexInTransaction: int64(op.Extend.OperationIndex),
				ExpiresAt:                   int64(op.Extend.NewExpiresAt),
			}

			log.Info("extend BTL", "params", extendParams)

			// Fetch the existing annotations before we update the entity, so that we
			// can re-insert them with the new block number.
			numericAnnotations, err := txDB.GetNumericAnnotations(ctx, sqlitegolem.GetNumericAnnotationsParams{
				EntityKey: strings.ToLower(op.Extend.EntityKey.Hex()),
				Block:     int64(blockWal.BlockInfo.Number),
			})
			if err != nil {
				return fmt.Errorf("failed to fetch annotations: %w", err)
			}

			stringAnnotations, err := txDB.GetStringAnnotations(ctx, sqlitegolem.GetStringAnnotationsParams{
				EntityKey: strings.ToLower(op.Extend.EntityKey.Hex()),
				Block:     int64(blockWal.BlockInfo.Number),
			})
			if err != nil {
				return fmt.Errorf("failed to fetch annotations: %w", err)
			}

			// Update the entity with the new expiry time
			err = txDB.UpdateEntityExpiresAt(ctx, extendParams)
			if err != nil {
				return fmt.Errorf("failed to extend entity BTL: %w", err)
			}

			for _, annotation := range numericAnnotations {
				value := uint64(annotation.Value)
				switch annotation.AnnotationKey {
				case arkivtype.SequenceAttributeKey:
					value = getSequence(
						blockWal.BlockInfo.Number,
						op.Extend.TransactionIndex,
						op.Extend.OperationIndex,
					)
				case arkivtype.ExpirationAttributeKey:
					value = op.Extend.NewExpiresAt
				}
				err = txDB.InsertNumericAnnotation(ctx, sqlitegolem.InsertNumericAnnotationParams{
					EntityKey:                         strings.ToLower(op.Extend.EntityKey.Hex()),
					EntityLastModifiedAtBlock:         int64(blockWal.BlockInfo.Number),
					EntityTransactionIndexInBlock:     int64(op.Extend.TransactionIndex),
					EntityOperationIndexInTransaction: int64(op.Extend.OperationIndex),
					AnnotationKey:                     annotation.AnnotationKey,
					Value:                             int64(value),
				})
				if err != nil {
					return fmt.Errorf("failed to insert numeric annotation: %w", err)
				}
			}

			for _, annotation := range stringAnnotations {
				err = txDB.InsertStringAnnotation(ctx, sqlitegolem.InsertStringAnnotationParams{
					EntityKey:                         strings.ToLower(op.Extend.EntityKey.Hex()),
					EntityLastModifiedAtBlock:         int64(blockWal.BlockInfo.Number),
					EntityTransactionIndexInBlock:     int64(op.Extend.TransactionIndex),
					EntityOperationIndexInTransaction: int64(op.Extend.OperationIndex),
					AnnotationKey:                     annotation.AnnotationKey,
					Value:                             annotation.Value,
				})
				if err != nil {
					return fmt.Errorf("failed to insert string annotation: %w", err)
				}
			}
		}

		marshalled, _ := json.Marshal(op)
		log.Info("operation", "operation", string(marshalled))
	}

	err = txDB.UpdateProcessingStatus(ctx, sqlitegolem.UpdateProcessingStatusParams{
		Network:                  networkID,
		LastProcessedBlockNumber: int64(blockWal.BlockInfo.Number),
		LastProcessedBlockHash:   blockWal.BlockInfo.Hash.Hex(),
	})
	if err != nil {
		return fmt.Errorf("failed to insert processing status: %w", err)
	}

	e.logDBStats("insertBlock", "before-commit")
	return tx.Commit()
}

var ErrStopIteration = errors.New("stop iteration")

func (e *SQLStore) QueryEntitiesInternalIterator(
	ctx context.Context,
	query string,
	args []any,
	options query.QueryOptions,
	iterator func(arkivtype.EntityData, arkivtype.Cursor) error,
) error {
	if e.databaseDisabled {
		return fmt.Errorf("database is disabled")
	}
	log.Info("Executing query", "query", query, "args", args)

	// Begin a read-only transaction for consistency
	e.logDBStats("queryEntities", "start")
	tx, err := e.readDB.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Safe to call even after commit

	_, err = tx.ExecContext(ctx, "PRAGMA temp_store = memory")
	if err != nil {
		return fmt.Errorf("failed to set temp store mode: %w", err)
	}

	txDB := sqlitegolem.New(tx)

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		if elapsed.Seconds() > 1 {
			e.logDBStats("queryPlan", "start")
			rows, err := e.readDB.QueryContext(context.Background(), fmt.Sprintf("explain query plan %s", query), args...)
			if err != nil {
				log.Error("failed to get query plan", "err", err)
				return
			}

			defer rows.Close()

			var (
				id      int
				parent  int
				notUsed int
				detail  string
			)

			b := strings.Builder{}
			for rows.Next() {
				err := rows.Err()
				if err != nil {
					log.Error("failed to get query plan", "err", err)
					return
				}

				err = rows.Scan(&id, &parent, &notUsed, &detail)
				if err != nil {
					log.Error("failed to get query plan", "err", err)
					return
				}
				fmt.Fprintf(&b, "id=%d parent=%d %s\n", id, parent, detail)
			}
			log.Info("query plan", "plan", b.String())
			e.logDBStats("queryPlan", "finish")
		}
	}()

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to get entities for query: %s: %w", query, err)
	}
	defer rows.Close()

	for rows.Next() {

		err := rows.Err()
		if err != nil {
			return fmt.Errorf("failed to get entities for query: %s: %w", query, err)
		}

		var (
			key                         *string
			expiresAt                   *uint64
			payload                     *[]byte
			contentType                 *string
			owner                       *string
			createdAtBlock              *uint64
			lastModifiedAtBlock         *uint64
			transactionIndexInBlock     *uint64
			operationIndexInTransaction *uint64
		)
		dest := []any{}
		columns := map[string]any{}
		for _, column := range options.AllColumns() {
			switch column {
			case "key":
				dest = append(dest, &key)
				columns[column] = &key
			case "expires_at":
				dest = append(dest, &expiresAt)
				columns[column] = &expiresAt
			case "payload":
				dest = append(dest, &payload)
				columns[column] = &payload
			case "content_type":
				dest = append(dest, &contentType)
				columns[column] = &contentType
			case "owner_address":
				dest = append(dest, &owner)
				columns[column] = &owner
			case "created_at_block":
				dest = append(dest, &createdAtBlock)
				columns[column] = &createdAtBlock
			case "last_modified_at_block":
				dest = append(dest, &lastModifiedAtBlock)
				columns[column] = &lastModifiedAtBlock
			case "transaction_index_in_block":
				dest = append(dest, &transactionIndexInBlock)
				columns[column] = &transactionIndexInBlock
			case "operation_index_in_transaction":
				dest = append(dest, &operationIndexInTransaction)
				columns[column] = &operationIndexInTransaction
			default:
				var value any
				dest = append(dest, &value)
				columns[column] = &value
			}
		}

		if err := rows.Scan(dest...); err != nil {
			return fmt.Errorf("failed to get entities for query: %s: %w", query, err)
		}

		var keyHash *common.Hash
		// We check whether the key was actually requested, since it's always included
		// in the query because of sorting
		if key != nil {
			hash := common.HexToHash(*key)
			keyHash = &hash
		}
		var value []byte
		if payload != nil {

			decoded, err := compression.BrotliDecompress(*payload)
			if err != nil {
				return fmt.Errorf("failed to decode compressed payload: %w", err)
			}

			value = decoded
		}
		var ownerAddress *common.Address
		if owner != nil {
			address := common.HexToAddress(*owner)
			ownerAddress = &address
		}

		r := arkivtype.EntityData{
			ExpiresAt:         expiresAt,
			Value:             value,
			ContentType:       contentType,
			Owner:             ownerAddress,
			CreatedAtBlock:    createdAtBlock,
			StringAttributes:  []entity.StringAnnotation{},
			NumericAttributes: []entity.NumericAnnotation{},
		}

		_, wantsKey := options.Columns["key"]
		if wantsKey {
			r.Key = keyHash
		}
		// Make sure to only include these properties when they were actually requested
		// They are always included in the query, so we need to explicitly check the query options
		_, wantsLastModified := options.Columns["last_modified_at_block"]
		if wantsLastModified {
			r.LastModifiedAtBlock = lastModifiedAtBlock
		}
		_, wantsTxIx := options.Columns["transaction_index_in_block"]
		if wantsTxIx {
			r.TransactionIndexInBlock = transactionIndexInBlock
		}
		_, wantsOpIx := options.Columns["operation_index_in_transaction"]
		if wantsOpIx {
			r.OperationIndexInTransaction = operationIndexInTransaction
		}

		cursor := arkivtype.Cursor{
			BlockNumber:  options.AtBlock,
			ColumnValues: make([]arkivtype.CursorValue, 0, len(options.OrderByColumns())),
		}

		for _, column := range options.OrderByColumns() {
			cursor.ColumnValues = append(cursor.ColumnValues, arkivtype.CursorValue{
				ColumnName: column.Name,
				Value:      columns[column.Name],
				Descending: column.Descending,
			})
		}

		if options.IncludeAnnotations {
			// Get string annotations
			stringAnnotRows, err := txDB.GetStringAnnotations(ctx, sqlitegolem.GetStringAnnotationsParams{
				EntityKey: strings.ToLower(keyHash.Hex()),
				Block:     int64(options.AtBlock),
			})
			if err != nil {
				return fmt.Errorf("failed to get string annotations: %w", err)
			}

			// Get numeric annotations
			numericAnnotRows, err := txDB.GetNumericAnnotations(ctx, sqlitegolem.GetNumericAnnotationsParams{
				EntityKey: strings.ToLower(keyHash.Hex()),
				Block:     int64(options.AtBlock),
			})
			if err != nil {
				return fmt.Errorf("failed to get numeric annotations: %w", err)
			}

			// Convert string annotations
			for _, row := range stringAnnotRows {
				if options.IncludeSyntheticAnnotations || !strings.HasPrefix(row.AnnotationKey, "$") {
					r.StringAttributes = append(r.StringAttributes, entity.StringAnnotation{
						Key:   row.AnnotationKey,
						Value: row.Value,
					})
				}
			}

			// Convert numeric annotations
			for _, row := range numericAnnotRows {
				if options.IncludeSyntheticAnnotations || !strings.HasPrefix(row.AnnotationKey, "$") {
					r.NumericAttributes = append(r.NumericAttributes, entity.NumericAnnotation{
						Key:   row.AnnotationKey,
						Value: uint64(row.Value),
					})
				}
			}
		}

		err = iterator(r, cursor)
		if errors.Is(err, ErrStopIteration) {
			break
		} else if err != nil {
			return fmt.Errorf("error during query execution: %w", err)
		}
	}

	// Commit the transaction (read-only, but ensures consistency)
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	e.logDBStats("queryEntities", "finish-commit")
	return nil
}
