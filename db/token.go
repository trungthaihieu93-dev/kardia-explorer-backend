// Package db
package db

import (
	"context"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

const (
	cToken = "Tokens"
)

type IToken interface {
	UpsertToken(ctx context.Context, t *types.Token) error
}

func (m *mongoDB) UpsertToken(ctx context.Context, t *types.Token) error {
	if _, err := m.wrapper.C(cContract).Insert(t); err != nil {
		return err
	}

	return nil
}
