// Package v2
package v2

import (
	"github.com/labstack/echo"

	"github.com/kardiachain/kardia-explorer-backend/api/e"
	"github.com/kardiachain/kardia-explorer-backend/api/private"
)

func bindPrivateAPIs(gr *echo.Group, h private.APIs) error {
	privateGr := gr.Group("/private")
	privateGr.Use(e.ValidatePrivateRequest())
	privateGr.POST("/staking/reload", func(c echo.Context) error {
		r := e.Response(c)
		if err := h.ReloadValidators(); err != nil {
			return r.Err(err)
		}
		return r.OK(nil)
	})
	privateGr.POST("/contracts/reload", func(c echo.Context) error {
		r := e.Response(c)

		return r.OK(nil)
	})
	privateGr.POST("/stats/reload", func(c echo.Context) error {
		r := e.Response(c)

		return r.OK(nil)
	})

	bindPrivateCrudAPIs(gr)
	return nil
}

func bindPrivateCrudAPIs(gr *echo.Group) {
	stakingGr := gr.Group("/staking")
	stakingGr.POST("/reload", func(c echo.Context) error {
		r := e.Response(c)

		return r.OK(nil)
	})
}
