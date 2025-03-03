package kardia

import (
	"context"

	kai "github.com/kardiachain/go-kardia"
	"github.com/kardiachain/go-kardia/rpc"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

// NewLogsFilter
func (ec *Client) NewLogsFilter(ctx context.Context, query kai.FilterQuery) (*rpc.ID, error) {
	return nil, nil
}

// UninstallFilter
func (ec *Client) UninstallFilter(ctx context.Context, filterID *rpc.ID) error {
	return nil
}

// GetFilterChanges
func (ec *Client) GetFilterChanges(ctx context.Context, filterID *rpc.ID) ([]*types.Log, error) {
	return nil, nil
}
