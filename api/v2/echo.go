// Package v2
package v2

import (
	"github.com/labstack/echo"

	"github.com/kardiachain/kardia-explorer-backend/api/private"
	"github.com/kardiachain/kardia-explorer-backend/api/public"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
)

func BindAPI(gr *echo.Group, cfg cfg.ExplorerConfig) error {
	v2Gr := gr.Group("/api/v2")
	privateConfig := private.Config{
		Logger: nil,
	}
	privateService, err := private.NewAPIs(privateConfig)
	if err != nil {
		return err
	}

	if err := bindPrivateAPIs(v2Gr, privateService); err != nil {
		return err
	}

	publicConfig := public.Config{
		Logger: nil,
	}
	publicService, err := public.NewAPIs(publicConfig)
	if err != nil {
		return err
	}
	if err := bindPublicAPIs(v2Gr, publicService); err != nil {
		return err
	}

	return nil
}
