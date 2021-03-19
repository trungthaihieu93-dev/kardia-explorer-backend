/*
 *  Copyright 2018 KardiaChain
 *  This file is part of the go-kardia library.
 *
 *  The go-kardia library is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Lesser General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  The go-kardia library is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU Lesser General Public License for more details.
 *
 *  You should have received a copy of the GNU Lesser General Public License
 *  along with the go-kardia library. If not, see <http://www.gnu.org/licenses/>.
 */

package api

import (
	"context"
	"fmt"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.uber.org/zap"

	v1 "github.com/kardiachain/kardia-explorer-backend/api/v1"
	"github.com/kardiachain/kardia-explorer-backend/api/v2"
	"github.com/kardiachain/kardia-explorer-backend/cache"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/server"
	"github.com/kardiachain/kardia-explorer-backend/utils"
)

func Start(cfg cfg.ExplorerConfig) {
	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.Logger())
	e.Use(middleware.Gzip())

	lgr, err := utils.NewLogger(cfg)
	if err != nil {
		panic("cannot init logger")
	}
	srvConfig := server.Config{
		StorageAdapter: db.Adapter(cfg.StorageDriver),
		StorageURI:     cfg.StorageURI,
		StorageDB:      cfg.StorageDB,
		StorageIsFlush: cfg.StorageIsFlush,

		KardiaURLs:         cfg.KardiaPublicNodes,
		KardiaTrustedNodes: cfg.KardiaTrustedNodes,

		CacheAdapter:      cache.Adapter(cfg.CacheEngine),
		CacheURL:          cfg.CacheURL,
		CacheDB:           cfg.CacheDB,
		CacheIsFlush:      cfg.CacheIsFlush,
		BlockBuffer:       cfg.BufferedBlocks,
		HttpRequestSecret: cfg.HttpRequestSecret,

		Metrics: nil,
		Logger:  lgr.With(zap.String("service", "APIs")),
	}
	srv, err := server.New(srvConfig)
	if err != nil {
		lgr.Panic("cannot create server instance %s", zap.Error(err))
	}
	ctx := context.Background()

	if cfg.IsReloadBootData {
		if err := srv.LoadBootContracts(ctx); err != nil {
			lgr.Panic("cannot load boot contracts", zap.Error(err))
		}
	}

	// Keep v1 APIs
	baseGr := e.Group("")
	if err := v1.BindAPI(baseGr, srv); err != nil {
		fmt.Println("cannot setup v1 APIs", err)
		panic(err)
	}

	if err := v2.BindAPI(baseGr, cfg); err != nil {
		fmt.Println("cannot setup v2 APIs", err)
		panic(err)
	}

	if err := e.Start(cfg.Port); err != nil {
		fmt.Println("cannot start echo server", err)
		panic(err)
	}
}
