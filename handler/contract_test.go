// Package handler
package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler_processContractInfo(t *testing.T) {
	ctx := context.Background()
	h, err := setupTestHandler()
	assert.Nil(t, err)
	_testTx := "0x330879ef1cf1bf5cb18205f218f3c1c80d5f2dece8edfc3222dc50b129b136f9"
	tx, err := h.w.TrustedNode().GetTransaction(ctx, _testTx)
	if err := h.processContractInfo(ctx, tx); err != nil {
		return
	}
}
