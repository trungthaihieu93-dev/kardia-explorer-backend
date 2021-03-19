// Package public
package public

import (
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cache"
	"github.com/kardiachain/kardia-explorer-backend/db"
)

type Config struct {
	Logger *zap.Logger
}

type APIs interface {
}

//service should not interact with node
type service struct {
	db     db.Client
	cache  cache.Client
	logger *zap.Logger
}

func NewAPIs(cfg Config) (APIs, error) {
	return &service{}, nil
}
