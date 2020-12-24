// Package db
package db

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/types"
)

type Adapter string

const (
	MGO Adapter = "mgo"
)

type Config struct {
	DbAdapter Adapter
	DbName    string
	URL       string
	MinConn   int
	MaxConn   int
	FlushDB   bool

	Logger *zap.Logger
}

// DB define list API used by infoServer
type Client interface {
	ping() error
	dropCollection(collectionName string)
	dropDatabase(ctx context.Context) error

	// Block details
	BlockByHeight(ctx context.Context, blockHeight uint64) (*types.Block, error)
	BlockByHash(ctx context.Context, blockHash string) (*types.Block, error)
	IsBlockExist(ctx context.Context, blockHeight uint64) (bool, error)
	BlockTxCount(ctx context.Context, hash string) (int64, error)

	// Interact with blocks
	Blocks(ctx context.Context, pagination *types.Pagination) ([]*types.Block, error)
	InsertBlock(ctx context.Context, block *types.Block) error
	UpsertBlock(ctx context.Context, block *types.Block) error
	DeleteLatestBlock(ctx context.Context) (uint64, error)
	// TODO(trinhdn): Replace delete+insert operation with upsert instead
	DeleteBlockByHeight(ctx context.Context, blockHeight uint64) error
	BlocksByProposer(ctx context.Context, proposer string, pagination *types.Pagination) ([]*types.Block, uint64, error)

	// Txs
	TxsCount(ctx context.Context) (uint64, error)
	TxsByBlockHash(ctx context.Context, blockHash string, pagination *types.Pagination) ([]*types.Transaction, uint64, error)
	TxsByBlockHeight(ctx context.Context, blockNumber uint64, pagination *types.Pagination) ([]*types.Transaction, uint64, error)
	TxsByAddress(ctx context.Context, address string, pagination *types.Pagination) ([]*types.Transaction, uint64, error)
	LatestTxs(ctx context.Context, pagination *types.Pagination) ([]*types.Transaction, error)

	// Tx detail
	TxByHash(ctx context.Context, txHash string) (*types.Transaction, error)
	TxByNonce(ctx context.Context, nonce int64) (*types.Transaction, error)

	// Interact with tx
	InsertTxs(ctx context.Context, txs []*types.Transaction) error
	UpsertTxs(ctx context.Context, txs []*types.Transaction) error
	InsertListTxByAddress(ctx context.Context, list []*types.TransactionByAddress) error

	// Token
	TokenHolders(ctx context.Context, tokenAddress string, pagination *types.Pagination) ([]*types.TokenHolder, uint64, error)
	//InternalTxs(ctx context.Context)

	// Address
	AddressByHash(ctx context.Context, addressHash string) (*types.Address, error)
	InsertAddress(ctx context.Context, address *types.Address) error
	OwnedTokensOfAddress(ctx context.Context, address string, pagination *types.Pagination) ([]*types.TokenHolder, uint64, error)

	// ActiveAddress
	UpdateAddresses(ctx context.Context, addresses map[string]*types.Address) error
	GetTotalAddresses(ctx context.Context) (uint64, uint64, error)
	GetListAddresses(ctx context.Context, sortDirection int, pagination *types.Pagination) ([]*types.Address, error)
}

func NewClient(cfg Config) (Client, error) {
	cfg.Logger.Debug("Create new db instance with config", zap.Any("config", cfg))
	switch cfg.DbAdapter {
	case MGO:
		return newMongoDB(cfg)
	default:
		return nil, errors.New("invalid db config")
	}
}
