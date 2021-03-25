// Package handler
package handler

import (
	"context"
	"fmt"

	ctypes "github.com/kardiachain/go-kardia/types"
	"go.uber.org/zap"
)

type ILogs interface {
	SubscribeLogs(ctx context.Context) error
}

func (h *handler) SubscribeLogs(ctx context.Context) error {
	lgr := h.logger.With(zap.String("method", "SubscribeValidatorEvent"))
	wsNode := h.w.WSNode()
	trustedNode := h.w.TrustedNode()
	nValidatorSMCAddresses, err := trustedNode.ValidatorSMCAddresses(ctx)
	if err != nil {
		lgr.Warn("cannot get validatorSMCAddresses", zap.Error(err))
	}
	var validatorsSMCAddresses []string
	for _, addr := range nValidatorSMCAddresses {
		validatorsSMCAddresses = append(validatorsSMCAddresses, addr.Hex())
	}
	headersCh := make(chan *ctypes.Header)
	sub, err := wsNode.SubscribeNewHead(context.Background(), headersCh)

	for {
		select {
		case err := <-sub.Err():
			lgr.Error("Subscribe error", zap.Error(err))
		case header := <-headersCh:
			h.processLogs(ctx, header)
		}
	}
}

func (h *handler) processLogs(ctx context.Context, header *ctypes.Header) {
	lgr := h.logger.With(zap.String("method", "processHeader"))
	block, err := h.w.TrustedNode().BlockByHash(ctx, header.Hash().Hex())
	if err != nil {
		lgr.Debug("cannot get block", zap.Error(err))
		return
	}
	if err := h.reloadProposer(ctx, block.ProposerAddress); err != nil {
		lgr.Error("cannot reload proposer", zap.Error(err))
	}

	if header.NumTxs == 0 {
		lgr.Debug("block has no txs", zap.String("hash", header.Hash().Hex()))
		return
	}

	//for _, tx := range block.Txs {
	//	//if tx.To == "0x" {
	//	//	if err := h.onContractCreation(ctx, tx); err != nil {
	//	//		lgr.Error("handle contract creation failed", zap.Error(err))
	//	//		return
	//	//	}
	//	//}
	//	//fmt.Printf("Tx detailt: %+v \n", tx)
	//}
	for _, r := range block.Receipts {
		fmt.Printf("Receipt: %+v \n", r)
	}

}
