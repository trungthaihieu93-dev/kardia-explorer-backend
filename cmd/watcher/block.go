// Package main
package main

import (
	"context"

	"github.com/kardiachain/kardia-explorer-backend/cache"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	//"github.com/kardiachain/kardia-explorer-backend/handler"
	blockHandler "github.com/kardiachain/kardia-explorer-backend/handler/block"

	"github.com/kardiachain/kardia-explorer-backend/utils"
)

func runBlockSubscriber(ctx context.Context, serviceCfg cfg.ExplorerConfig) error {

	logger, err := utils.NewLogger(serviceCfg)
	if err != nil {
		panic(err.Error())
	}

	handlerCfg := blockHandler.Config{
		TrustedNodes: serviceCfg.KardiaTrustedNodes,
		PublicNodes:  serviceCfg.KardiaPublicNodes,
		WSNodes:      serviceCfg.KardiaWSNodes,

		StorageAdapter: db.Adapter(serviceCfg.StorageDriver),
		StorageURI:     serviceCfg.StorageURI,
		StorageDB:      serviceCfg.StorageDB,

		CacheAdapter: cache.Adapter(serviceCfg.CacheEngine),
		CacheURL:     serviceCfg.CacheURL,
		CacheDB:      serviceCfg.CacheDB,

		Logger: logger,
	}
	h, err := blockHandler.New(handlerCfg)
	if err != nil {
		return err
	}

	go h.SubscribeNewBlock(ctx)
	return nil
}
