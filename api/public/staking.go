// Package public
package public

import (
	"context"

	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/types"
)

type Staking interface {
	StakingStats(ctx context.Context) (*types.StakingStats, error)
	Validators(ctx context.Context, filter ValidatorFilter)
}

func (s *service) StakingStats(ctx context.Context) (*types.StakingStats, error) {
	return s.cache.StakingStats(ctx)
}

func (s *service) Validators(ctx context.Context, filter ValidatorFilter) ([]*types.Validator, error) {
	dbFilter := db.ValidatorsFilter{
		Role: 0,
		Skip: 0,
	}
	validators, err := s.db.Validators(ctx, dbFilter)
	if err != nil {
		return nil, err
	}
	return validators, nil
}

func (s *service) Validator() (*types.Validator, error) {
	return nil, nil
}

func (s *service) ValidatorsOfDelegator() ([]*types.ValidatorsByDelegator, error) {
	return nil, nil
}
