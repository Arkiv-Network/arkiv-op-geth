package eth

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/golem-base/arkivtype"
	"github.com/ethereum/go-ethereum/golem-base/query"
	"github.com/ethereum/go-ethereum/golem-base/sqlstore"
	"github.com/ethereum/go-ethereum/golem-base/storageaccounting"
	"github.com/ethereum/go-ethereum/log"
)

type IncludeData struct {
	Key                         bool `json:"key"`
	Attributes                  bool `json:"attributes"`
	Payload                     bool `json:"payload"`
	ContentType                 bool `json:"contentType"`
	Expiration                  bool `json:"expiration"`
	Owner                       bool `json:"owner"`
	CreatedAtBlock              bool `json:"createdAtBlock"`
	LastModifiedAtBlock         bool `json:"lastModifiedAtBlock"`
	TransactionIndexInBlock     bool `json:"transactionIndexInBlock"`
	OperationIndexInTransaction bool `json:"operationIndexInTransaction"`
}

type QueryOptions struct {
	AtBlock        *uint64                       `json:"atBlock"`
	IncludeData    *IncludeData                  `json:"includeData"`
	OrderBy        []arkivtype.OrderByAnnotation `json:"orderBy"`
	ResultsPerPage uint64                        `json:"resultsPerPage"`
	Cursor         string                        `json:"cursor"`
}

var defaultColumns = []string{
	arkivtype.GetColumnOrPanic("key"),
	arkivtype.GetColumnOrPanic("expires_at"),
	arkivtype.GetColumnOrPanic("owner_address"),
	arkivtype.GetColumnOrPanic("payload"),
	arkivtype.GetColumnOrPanic("content_type"),
}

func (options *QueryOptions) toInternalQueryOptions() (*internalQueryOptions, error) {
	switch {
	case options == nil:
		return &internalQueryOptions{
			Columns:            defaultColumns,
			IncludeAnnotations: true,
		}, nil
	case options.IncludeData == nil:
		return &internalQueryOptions{
			Columns:            defaultColumns,
			IncludeAnnotations: true,
			OrderBy:            options.OrderBy,
			AtBlock:            options.AtBlock,
			Cursor:             options.Cursor,
		}, nil
	default:
		iq := internalQueryOptions{
			Columns: []string{},
			OrderBy: options.OrderBy,
			AtBlock: options.AtBlock,
			Cursor:  options.Cursor,
		}
		if options.IncludeData.Attributes {
			iq.IncludeAnnotations = true
		}
		if options.IncludeData.Payload {
			iq.Columns = append(iq.Columns, arkivtype.GetColumnOrPanic("payload"))
		}
		if options.IncludeData.ContentType {
			iq.Columns = append(iq.Columns, arkivtype.GetColumnOrPanic("content_type"))
		}
		if options.IncludeData.Expiration {
			iq.Columns = append(iq.Columns, arkivtype.GetColumnOrPanic("expires_at"))
		}
		if options.IncludeData.Owner {
			iq.Columns = append(iq.Columns, arkivtype.GetColumnOrPanic("owner_address"))
		}
		if options.IncludeData.Key {
			iq.Columns = append(iq.Columns, arkivtype.GetColumnOrPanic("key"))
		}
		if options.IncludeData.CreatedAtBlock {
			iq.Columns = append(iq.Columns, arkivtype.GetColumnOrPanic("created_at_block"))
		}
		if options.IncludeData.LastModifiedAtBlock {
			iq.Columns = append(iq.Columns, arkivtype.GetColumnOrPanic("last_modified_at_block"))
		}
		if options.IncludeData.TransactionIndexInBlock {
			iq.Columns = append(iq.Columns, arkivtype.GetColumnOrPanic("transaction_index_in_block"))
		}
		if options.IncludeData.OperationIndexInTransaction {
			iq.Columns = append(iq.Columns, arkivtype.GetColumnOrPanic("operation_index_in_transaction"))
		}
		return &iq, nil
	}
}

type internalQueryOptions struct {
	AtBlock            *uint64                       `json:"atBlock"`
	IncludeAnnotations bool                          `json:"includeAnnotations"`
	Columns            []string                      `json:"columns"`
	OrderBy            []arkivtype.OrderByAnnotation `json:"orderBy"`
	Cursor             string                        `json:"cursor"`
}

type arkivAPI struct {
	eth   *Ethereum
	store *sqlstore.SQLStore
}

func NewArkivAPI(eth *Ethereum, store *sqlstore.SQLStore) *arkivAPI {
	return &arkivAPI{
		eth:   eth,
		store: store,
	}
}

func (api *arkivAPI) Query(
	ctx context.Context,
	req string,
	op *QueryOptions,
) (*arkivtype.QueryResponse, error) {

	log.Info("query options", "options", op)

	expr, err := query.Parse(req)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	options, err := op.toInternalQueryOptions()
	if err != nil {
		return nil, err
	}

	latestsHead := api.eth.blockchain.CurrentBlock()
	block := latestsHead.Number.Uint64()

	queryOptions := query.QueryOptions{
		IncludeAnnotations: options.IncludeAnnotations,
		Columns:            options.Columns,
		OrderBy:            options.OrderBy,
	}

	if len(options.Cursor) != 0 {
		offset, err := queryOptions.DecodeCursor(options.Cursor)
		if err != nil {
			return nil, err
		}
		block = offset.BlockNumber
		queryOptions.Cursor = offset.ColumnValues
	}

	if options.AtBlock != nil {
		block = *options.AtBlock
	}

	queryOptions.AtBlock = block

	query, err := expr.Evaluate(&queryOptions)
	if err != nil {
		return nil, err
	}

	response := &arkivtype.QueryResponse{
		BlockNumber: block,
		Data:        make([]json.RawMessage, 0),
	}

	// In case the query should be run on a block that we don't have yet,
	// we wait for a little bit to see if we receive the block.
	if block > latestsHead.Number.Uint64() {
		var cadence time.Duration
		if latestsHead.Number.Uint64() <= 1 {
			// For block 1, we cannot deduce the cadence, so we just assume 2 seconds
			cadence = 2 * time.Second
		} else {
			cadence = time.Duration(
				latestsHead.Time-api.eth.blockchain.GetHeaderByHash(latestsHead.ParentHash).Time,
			) * time.Second
		}

		time.Sleep(2 * time.Duration(cadence) * time.Second)
		latestsHead = api.eth.blockchain.CurrentBlock()
		if block > latestsHead.Number.Uint64() {
			return nil, fmt.Errorf("requested block is in the future")
		}
	}

	// 256 bytes is for the overhead of the 'envelope' around the entity data
	// and the separator characters in between
	responseSize := 256

	// 512 kb
	maxResponseSize := 512 * 1024 * 1024
	maxResultsPerPage := 0

	if op != nil {
		maxResultsPerPage = int(op.ResultsPerPage)
		log.Info("query max results per page", "value", maxResultsPerPage)
	}

	startTime := time.Now()

	defer func() {
		elapsed := time.Since(startTime)
		log.Info("query execution time", "elapsed", elapsed)
	}()

	err = api.store.QueryEntitiesInternalIterator(
		ctx,
		query.Query,
		query.Args,
		queryOptions,
		func(entity arkivtype.EntityData, cursor arkivtype.Cursor) error {

			ed, err := json.Marshal(entity)
			if err != nil {
				return fmt.Errorf("failed to marshal entity: %w", err)
			}

			newLen := responseSize + len(ed) + 1
			if newLen > maxResponseSize {
				cursor, err := queryOptions.EncodeCursor(&cursor)
				if err != nil {
					return fmt.Errorf("could not encode offset: %w", err)
				}
				response.Cursor = &cursor
				return sqlstore.ErrStopIteration
			}
			response.Data = append(response.Data, ed)

			if maxResultsPerPage > 0 && len(response.Data) >= maxResultsPerPage {
				cursor, err := queryOptions.EncodeCursor(&cursor)
				if err != nil {
					return fmt.Errorf("could not encode offset: %w", err)
				}
				response.Cursor = &cursor
				return sqlstore.ErrStopIteration
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	log.Info("query number of results", "value", len(response.Data))
	return response, nil
}

// GetEntityCount returns the total number of entities in the storage.
func (api *arkivAPI) GetEntityCount(ctx context.Context) (uint64, error) {
	count, err := api.store.GetEntityCount(ctx, api.eth.blockchain.CurrentBlock().Number.Uint64())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (api *arkivAPI) GetNumberOfUsedSlots() (*hexutil.Big, error) {
	header := api.eth.blockchain.CurrentBlock()
	stateDB, err := api.eth.BlockChain().StateAt(header.Root)
	if err != nil {
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	counter := storageaccounting.GetNumberOfUsedSlots(stateDB)
	counterAsBigInt := big.NewInt(0)
	counter.IntoBig(&counterAsBigInt)
	return (*hexutil.Big)(counterAsBigInt), nil
}

type BlockTiming struct {
	CurrentBlock     uint64 `json:"current_block"`
	CurrentBlockTime uint64 `json:"current_block_time"`
	BlockDuration    uint64 `json:"duration"`
}

func (api *arkivAPI) GetBlockTiming(ctx context.Context) (*BlockTiming, error) {
	header := api.eth.blockchain.CurrentHeader()
	previousHeader := api.eth.blockchain.GetHeaderByHash(header.ParentHash)
	if previousHeader == nil {
		return nil, fmt.Errorf("failed to get previous header")
	}

	return &BlockTiming{
		CurrentBlock:     header.Number.Uint64(),
		CurrentBlockTime: header.Time,
		BlockDuration:    header.Time - previousHeader.Time,
	}, nil
}
