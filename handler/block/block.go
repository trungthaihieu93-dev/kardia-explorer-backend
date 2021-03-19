// Package block
package block

import (
	"context"

	ktypes "github.com/kardiachain/go-kaiclient/kardia"
	ctypes "github.com/kardiachain/go-kardia/types"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cache"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/kardia"
	"github.com/kardiachain/kardia-explorer-backend/types"
)

type Config struct {
	TrustedNodes []string
	PublicNodes  []string
	WSNodes      []string

	// DB config
	StorageAdapter db.Adapter
	StorageURI     string
	StorageDB      string

	CacheAdapter cache.Adapter
	CacheURL     string
	CacheDB      int

	Logger *zap.Logger
}

type Handler interface {
	SubscribeNewBlock(ctx context.Context) error
}

type handler struct {
	// Internal
	w      *kardia.Wrapper
	db     db.Client
	cache  cache.Client
	logger *zap.Logger
}

func New(cfg Config) (Handler, error) {
	wrapperCfg := kardia.WrapperConfig{
		TrustedNodes: cfg.TrustedNodes,
		PublicNodes:  cfg.PublicNodes,
		WSNodes:      cfg.WSNodes,
		Logger:       cfg.Logger,
	}
	kardiaWrapper, err := kardia.NewWrapper(wrapperCfg)
	if err != nil {
		return nil, err
	}

	dbConfig := db.Config{
		DbAdapter: cfg.StorageAdapter,
		DbName:    cfg.StorageDB,
		URL:       cfg.StorageURI,
		Logger:    cfg.Logger,
		MinConn:   1,
		MaxConn:   4,
	}
	dbClient, err := db.NewClient(dbConfig)
	if err != nil {
		return nil, err
	}

	cacheCfg := cache.Config{
		Adapter: cfg.CacheAdapter,
		URL:     cfg.CacheURL,
		DB:      cfg.CacheDB,
		Logger:  cfg.Logger,
	}
	cacheClient, err := cache.New(cacheCfg)
	if err != nil {
		return nil, err
	}

	return &handler{
		w:      kardiaWrapper,
		logger: cfg.Logger.With(zap.String("handler", "newBlock")),
		db:     dbClient,
		cache:  cacheClient,
	}, nil
}

func (h *handler) SubscribeNewBlock(ctx context.Context) error {
	lgr := h.logger.With(zap.String("method", "SubscribeValidatorEvent"))
	wsNode := h.w.WSNode()
	trustedNode := h.w.TrustedNode()
	nValidatorSMCAddresses, err := trustedNode.ValidatorSMCAddresses(ctx)
	if err != nil {
		lgr.Warn("cannot get validatorSMCAddresses", zap.Error(err))
	}
	var validatorsSMCAddresses []string
	for _, addr := range nValidatorSMCAddresses {
		validatorsSMCAddresses = append(validatorsSMCAddresses, addr.Hex())
	}
	headersCh := make(chan *ctypes.Header)
	sub, err := wsNode.SubscribeNewHead(context.Background(), headersCh)

	for {
		select {
		case err := <-sub.Err():
			lgr.Error("Subscribe error", zap.Error(err))
		case header := <-headersCh:
			h.processHeader(ctx, header)
		}
	}
}

func (h *handler) processHeader(ctx context.Context, header *ctypes.Header) {
	lgr := h.logger.With(zap.String("method", "processHeader"))
	nBlock, err := h.w.TrustedNode().BlockByHeight(ctx, header.Height)
	if err != nil {
		lgr.Debug("cannot get block", zap.Error(err))
		return
	}

	block := SanitizeBlock(nBlock)

	lgr.Info("Importing block:", zap.Uint64("Height", block.Height),
		zap.Int("Txs length", len(block.Txs)), zap.Int("Receipts length", len(block.Receipts)))
	if isExist, err := h.db.IsBlockExist(ctx, block.Height); err != nil || isExist {
		lgr.Error("cannot found block", zap.Error(err))
		return
	}
	if err := h.cache.InsertBlock(ctx, block); err != nil {
		lgr.Debug("cannot import block to cache", zap.Error(err))
	}

	if err := h.db.InsertBlock(ctx, block); err != nil {
		lgr.Error("cannot insert block", zap.Error(err))
		return
	}
}

func SanitizeBlock(b *ktypes.Block) *types.Block {
	return &types.Block{
		Hash:            b.Hash,
		Height:          b.Height,
		CommitHash:      b.CommitHash,
		GasLimit:        b.GasLimit,
		GasUsed:         b.GasUsed,
		Rewards:         b.Rewards,
		NumTxs:          b.NumTxs,
		Time:            b.Time,
		ProposerAddress: b.ProposerAddress,
		LastBlock:       b.LastBlock,
		DataHash:        b.DataHash,
		ReceiptsRoot:    b.ReceiptsRoot,
		//LogsBloom:         b.LogsBloom,
		ValidatorHash:     b.ValidatorHash,
		NextValidatorHash: b.NextValidatorHash,
		ConsensusHash:     b.ConsensusHash,
		AppHash:           b.AppHash,
		EvidenceHash:      b.EvidenceHash,
		//NumDualEvents:     b.Du,
		//DualEventsHash:    b.D,
		//Txs:               nil,
		//Receipts:          nil,
	}
}
