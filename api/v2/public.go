// Package v2
package v2

import (
	"github.com/labstack/echo"

	"github.com/kardiachain/kardia-explorer-backend/api/public"
)

func bindPublicAPIs(gr *echo.Group, h public.APIs) error {
	// Public should use direct group

	return nil
}

func bindDashboardAPIs(gr *echo.Group, h public.APIs) {

}
