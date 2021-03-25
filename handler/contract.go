// Package handler
package handler

import (
	"context"
	"fmt"

	"github.com/kardiachain/go-kaiclient/kardia"
	ctypes "github.com/kardiachain/go-kardia/types"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/types"
)

type IContractHandler interface {
	SubscribeContractCreationEvent(ctx context.Context) error
}

// SubscribeContractCreationEvent subscribe all tx related to ContractCreation
func (h *handler) SubscribeContractCreationEvent(ctx context.Context) error {
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
			h.onNewContractEvent(ctx, header)
		}
	}
}

func (h *handler) onNewContractEvent(ctx context.Context, header *ctypes.Header) {
	lgr := h.logger.With(zap.String("method", "onNewContract"))
	lgr.Debug("New Header", zap.Any("h", header))
	if header.NumTxs == 0 {
		return
	}
	block, err := h.w.TrustedNode().BlockByHash(ctx, header.Hash().Hex())
	if err != nil {
		lgr.Debug("cannot get block", zap.Error(err))
		return
	}

	for _, tx := range block.Txs {
		if tx.To != cfg.CreatorAddress {
			continue
		}
		if err := h.processContractInfo(ctx, tx); err != nil {
			lgr.Error("cannot process contract info", zap.Error(err))
			return
		}
	}
}

func (h *handler) processContractInfo(ctx context.Context, tx *kardia.Transaction) error {
	lgr := h.logger.With(zap.String("method", "processContractInfo"))
	node := h.w.TrustedNode()

	receipt, err := node.GetTransactionReceipt(ctx, tx.Hash)
	if err != nil {
		lgr.Error("cannot get transaction receipt", zap.Error(err))
		return err
	}

	// Failed
	if receipt.Status == 0 {
		lgr.Debug("create new contract failed, so ignore")
		return fmt.Errorf("create new contract failed")
	}
	l := receipt.Logs[0]
	contractAddress := l.Address
	fmt.Println("ContractAddress", contractAddress)

	// Insert as normal contract
	ownerAddress := tx.From
	createAtBlock := tx.BlockNumber
	txHash := tx.Hash
	byteCode := tx.InputData
	dbContract := &types.Contract{
		Address:       contractAddress,
		Bytecode:      byteCode,
		OwnerAddress:  ownerAddress,
		TxHash:        txHash,
		CreateAtBlock: createAtBlock,
		Type:          types.ContractTypeDefault,
		Status:        types.ContractStatusUnverified,
	}
	if err := h.db.UpsertContract(ctx, dbContract); err != nil {
		lgr.Error("cannot upsert contract", zap.Error(err))
		return err
	}

	return nil
}

func (h *handler) onCreateKRC20Token() {

}
