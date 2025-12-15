package eth

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"

	queryapi "github.com/Arkiv-Network/query-api/query"
	sqlitestore "github.com/Arkiv-Network/sqlite-store"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/golem-base/storageaccounting"
	"github.com/ethereum/go-ethereum/log"
)

type arkivAPI struct {
	eth   *Ethereum
	store *sqlitestore.SQLiteStore
}

func NewArkivAPI(eth *Ethereum, dbFile string) (*arkivAPI, error) {
	logger := slog.New(log.Root().Handler())
	store, err := sqlitestore.NewSQLiteStore(logger, dbFile)
	if err != nil {
		return nil, fmt.Errorf("error creating sqlite store: %w", err)
	}
	return &arkivAPI{
		eth:   eth,
		store: store,
	}, nil
}

func (api *arkivAPI) Query(
	ctx context.Context,
	req string,
	op *queryapi.Options,
) (*queryapi.QueryResponse, error) {

	lastBlock := api.eth.blockchain.CurrentHeader().Number.Uint64()

	log.Info("api", "last_block", lastBlock)

	if op == nil {
		op = &queryapi.Options{}
	}
	if op.AtBlock == nil {
		op.AtBlock = &lastBlock
	}

	response, err := api.store.QueryEntities(ctx, req, op, "sqlite")
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}

	return response, nil
}

// GetEntityCount returns the total number of entities in the storage.
func (api *arkivAPI) GetEntityCount(ctx context.Context) (uint64, error) {

	return 0, nil
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
