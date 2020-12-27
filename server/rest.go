// Package server
package server

import (
	"context"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/kardiachain/explorer-backend/cfg"

	"github.com/kardiachain/go-kardia/lib/common"

	"github.com/bxcodec/faker/v3"
	"github.com/labstack/echo"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/api"
	"github.com/kardiachain/explorer-backend/types"
)

func (s *Server) Ping(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) Info(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) Stats(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	blocks, err := s.dbClient.Blocks(ctx, &types.Pagination{
		Skip:  0,
		Limit: 11,
	})
	if err != nil {
		return api.InternalServer.Build(c)
	}

	type Stat struct {
		NumTxs uint64 `json:"numTxs"`
		Time   uint64 `json:"time"`
	}

	var stats []*Stat
	for _, b := range blocks {
		stat := &Stat{
			NumTxs: b.NumTxs,
			Time:   uint64(b.Time.Unix()),
		}
		stats = append(stats, stat)
	}

	return api.OK.SetData(struct {
		Data interface{} `json:"data"`
	}{
		Data: stats,
	}).Build(c)
}

func (s *Server) TotalHolders(c echo.Context) error {
	ctx := context.Background()
	totalHolders, totalContracts := s.cacheClient.TotalHolders(ctx)
	return api.OK.SetData(struct {
		TotalHolders   uint64 `json:"totalHolders"`
		TotalContracts uint64 `json:"totalContracts"`
	}{
		TotalHolders:   totalHolders,
		TotalContracts: totalContracts,
	}).Build(c)
}

func (s *Server) Nodes(c echo.Context) error {
	ctx := context.Background()
	nodes, err := s.kaiClient.NodesInfo(ctx)
	if err != nil {
		s.logger.Warn("cannot get nodes info from RPC", zap.Error(err))
		return api.Invalid.Build(c)
	}
	var result []*NodeInfo
	for _, node := range nodes {
		result = append(result, &NodeInfo{
			ID:         node.ID,
			Moniker:    node.Moniker,
			PeersCount: len(node.Peers),
		})
	}
	return api.OK.SetData(result).Build(c)
}

func (s *Server) TokenInfo(c echo.Context) error {
	ctx := context.Background()
	if !s.cacheClient.IsRequestToCoinMarket(ctx) {
		tokenInfo, err := s.cacheClient.TokenInfo(ctx)
		if err != nil {
			tokenInfo, err = s.infoServer.TokenInfo(ctx)
			if err != nil {
				return api.Invalid.Build(c)
			}
		}
		return api.OK.SetData(tokenInfo).Build(c)
	}

	tokenInfo, err := s.infoServer.TokenInfo(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}
	return api.OK.SetData(tokenInfo).Build(c)
}

func (s *Server) UpdateSupplyAmounts(c echo.Context) error {
	ctx := context.Background()
	if !strings.Contains(c.Request().Header.Get("Authorization"), s.infoServer.HttpRequestSecret) {
		return api.Unauthorized.Build(c)
	}
	var supplyInfo *types.SupplyInfo
	if err := c.Bind(&supplyInfo); err != nil {
		return api.Invalid.Build(c)
	}
	if err := s.cacheClient.UpdateSupplyAmounts(ctx, supplyInfo); err != nil {
		return api.Invalid.Build(c)
	}
	return api.OK.Build(c)
}

func (s *Server) TPS(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) ValidatorStats(c echo.Context) error {
	ctx := context.Background()
	var (
		page, limit int
		err         error
	)
	pagination, page, limit := getPagingOption(c)

	// get validators list from cache
	valsList, err := s.cacheClient.Validators(ctx)
	if err != nil {
		s.logger.Debug("cannot get validators list from cache", zap.Error(err))
		valsList, err = s.getValidatorsList(ctx)
		if err != nil {
			return api.Invalid.Build(c)
		}
	}

	// get delegation details
	validator, err := s.kaiClient.Validator(ctx, c.Param("address"))
	if err != nil {
		s.logger.Warn("cannot get validator info from RPC, use cached validator info instead", zap.Error(err))
	}
	// get validator additional info such as commission rate
	for _, val := range valsList.Validators {
		if strings.ToLower(val.Address.Hex()) == strings.ToLower(c.Param("address")) {
			if validator == nil {
				validator = val
				break
			}
			// update validator VotingPowerPercentage
			validator.VotingPowerPercentage = val.VotingPowerPercentage
			break
		}
	}
	if validator == nil {
		// address in param is not a validator
		return api.Invalid.Build(c)
	}
	var delegators []*types.Delegator
	if pagination.Skip > len(validator.Delegators) {
		delegators = []*types.Delegator(nil)
	} else if pagination.Skip+pagination.Limit > len(validator.Delegators) {
		delegators = validator.Delegators[pagination.Skip:len(validator.Delegators)]
	} else {
		delegators = validator.Delegators[pagination.Skip : pagination.Skip+pagination.Limit]
	}

	total := uint64(len(validator.Delegators))
	validator.Delegators = delegators

	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  validator,
	}).Build(c)
}

func (s *Server) Validators(c echo.Context) error {
	ctx := context.Background()
	valsList, err := s.getValidatorsList(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}
	var result []*types.Validator
	for _, val := range valsList.Validators {
		if val.Role != 0 {
			result = append(result, val)
		}
	}
	valsList.Validators = result
	return api.OK.SetData(valsList).Build(c)
}

func (s *Server) GetValidatorsByDelegator(c echo.Context) error {
	ctx := context.Background()
	delAddr := c.Param("address")
	valsList, err := s.kaiClient.GetValidatorsByDelegator(ctx, common.HexToAddress(delAddr))
	if err != nil {
		return api.Invalid.Build(c)
	}
	return api.OK.SetData(valsList).Build(c)
}

func (s *Server) GetCandidatesList(c echo.Context) error {
	ctx := context.Background()
	valsList, err := s.getValidatorsList(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}
	var (
		result    []*types.Validator
		valsCount = 0
	)
	for _, val := range valsList.Validators {
		if val.Role == 0 {
			result = append(result, val)
		} else {
			valsCount++
		}
	}
	valsList.Validators = result
	return api.OK.SetData(valsList).Build(c)
}

func (s *Server) GetSlashEvents(c echo.Context) error {
	ctx := context.Background()
	slashEvents, err := s.kaiClient.GetSlashEvents(ctx, common.HexToAddress(c.Param("address")))
	if err != nil {
		s.logger.Warn("Cannot GetSlashEvents", zap.Error(err))
		return api.Invalid.Build(c)
	}
	return api.OK.SetData(slashEvents).Build(c)
}

func (s *Server) GetSlashedTokens(c echo.Context) error {
	ctx := context.Background()
	result, err := s.kaiClient.GetTotalSlashedToken(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}
	return api.OK.SetData(result).Build(c)
}

func (s *Server) Blocks(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var (
		err    error
		blocks []*types.Block
	)
	pagination, page, limit := getPagingOption(c)

	// todo @londnd: implement read from cache,
	blocks, err = s.cacheClient.LatestBlocks(ctx, pagination)
	if err != nil || blocks == nil {
		s.logger.Debug("Cannot get latest blocks from cache", zap.Error(err))
		blocks, err = s.dbClient.Blocks(ctx, pagination)
		if err != nil {
			s.logger.Debug("Cannot get latest blocks from db", zap.Error(err))
			return api.InternalServer.Build(c)
		}
		s.logger.Debug("Got latest blocks from db")
	} else {
		s.logger.Debug("Got latest blocks from cache")
	}

	smcAddress := s.getValidatorsAddressAndRole(ctx)
	var result Blocks
	for _, block := range blocks {
		b := SimpleBlock{
			Height:          block.Height,
			Hash:            block.Hash,
			Time:            block.Time,
			ProposerAddress: block.ProposerAddress,
			NumTxs:          block.NumTxs,
			GasLimit:        block.GasLimit,
			GasUsed:         block.GasUsed,
			Rewards:         block.Rewards,
		}
		b.ProposerName = smcAddress[b.ProposerAddress].Name
		result = append(result, b)
	}
	total := s.cacheClient.LatestBlockHeight(ctx)
	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Data:  result,
		Total: total,
	}).Build(c)
}

func (s *Server) Block(c echo.Context) error {
	ctx := context.Background()
	blockHashOrHeightStr := c.Param("block")
	var (
		block *types.Block
		err   error
	)
	if strings.HasPrefix(blockHashOrHeightStr, "0x") {
		// get block in cache if exist
		block, err = s.cacheClient.BlockByHash(ctx, blockHashOrHeightStr)
		if err != nil {
			s.logger.Debug("got block by hash from cache error", zap.Any("block", block), zap.Error(err))
			// otherwise, get from db
			block, err = s.dbClient.BlockByHash(ctx, blockHashOrHeightStr)
			if err != nil {
				s.logger.Warn("got block by hash from db error", zap.Any("block", block), zap.Error(err))
				// try to get from RPC at last
				block, err = s.kaiClient.BlockByHash(ctx, blockHashOrHeightStr)
				if err != nil {
					s.logger.Warn("got block by hash from RPC error", zap.Any("block", block), zap.Error(err))
					return api.Invalid.Build(c)
				}
				s.logger.Info("got block by hash from RPC:", zap.Any("block", block), zap.Error(err))
			} else {
				s.Logger.Info("got block by hash from db:", zap.String("blockHash", blockHashOrHeightStr))
			}
		} else {
			s.Logger.Info("got block by hash from cache:", zap.String("blockHash", blockHashOrHeightStr))
		}
	} else {
		blockHeight, err := strconv.ParseUint(blockHashOrHeightStr, 10, 64)
		if err != nil || blockHeight <= 0 {
			return api.Invalid.Build(c)
		}
		// get block in cache if exist
		block, err = s.cacheClient.BlockByHeight(ctx, blockHeight)
		if err != nil {
			s.logger.Debug("got block by height from cache error", zap.Uint64("blockHeight", blockHeight), zap.Error(err))
			// otherwise, get from db
			block, err = s.dbClient.BlockByHeight(ctx, blockHeight)
			if err != nil {
				s.logger.Warn("got block by height from db error", zap.Uint64("blockHeight", blockHeight), zap.Error(err))
				// try to get from RPC at last
				block, err = s.kaiClient.BlockByHeight(ctx, blockHeight)
				if err != nil {
					s.logger.Warn("got block by height from RPC error", zap.Uint64("blockHeight", blockHeight), zap.Error(err))
					return api.Invalid.Build(c)
				}
				s.logger.Info("got block by height from RPC:", zap.Uint64("blockHeight", blockHeight), zap.Error(err))
			}
			s.Logger.Info("got block by height from db:", zap.Uint64("blockHeight", blockHeight))
		} else {
			s.Logger.Info("got block by height from cache:", zap.Uint64("blockHeight", blockHeight))
		}
	}
	smcAddress := s.getValidatorsAddressAndRole(ctx)
	result := &Block{
		Block:        *block,
		ProposerName: smcAddress[block.ProposerAddress].Name,
	}
	return api.OK.SetData(result).Build(c)
}

func (s *Server) PersistentErrorBlocks(c echo.Context) error {
	ctx := context.Background()
	heights, err := s.cacheClient.PersistentErrorBlockHeights(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}
	return api.OK.SetData(heights).Build(c)
}

func (s *Server) BlockExist(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) BlockTxs(c echo.Context) error {
	ctx := context.Background()
	block := c.Param("block")
	pagination, page, limit := getPagingOption(c)

	var (
		txs   []*types.Transaction
		total uint64
		err   error
	)

	if strings.HasPrefix(block, "0x") {
		// get block txs in block if exist
		txs, total, err = s.cacheClient.TxsByBlockHash(ctx, block, pagination)
		if err != nil {
			s.logger.Debug("cannot get block txs by hash from cache", zap.String("blockHash", block), zap.Error(err))
			// otherwise, get from db
			txs, total, err = s.dbClient.TxsByBlockHash(ctx, block, pagination)
			if err != nil {
				s.logger.Warn("cannot get block txs by hash from db", zap.String("blockHash", block), zap.Error(err))
				// try to get block txs from RPC
				blockRPC, err := s.kaiClient.BlockByHash(ctx, block)
				if err != nil {
					s.logger.Warn("cannot get block txs by hash from RPC", zap.String("blockHash", block), zap.Error(err))
					return api.InternalServer.Build(c)
				}
				txs = blockRPC.Txs
				if pagination.Skip > len(txs) {
					txs = []*types.Transaction(nil)
				} else if pagination.Skip+pagination.Limit > len(txs) {
					txs = blockRPC.Txs[pagination.Skip:len(txs)]
				} else {
					txs = blockRPC.Txs[pagination.Skip : pagination.Skip+pagination.Limit]
				}
				total = blockRPC.NumTxs
				s.Logger.Debug("got block txs by hash from RPC:", zap.String("blockHash", block))
			} else {
				s.Logger.Debug("got block txs by hash from db:", zap.String("blockHash", block))
			}
		} else {
			s.Logger.Debug("got block txs by hash from cache:", zap.String("blockHash", block))
		}
	} else {
		height, err := strconv.ParseUint(block, 10, 64)
		if err != nil || height <= 0 {
			return api.Invalid.Build(c)
		}
		// get block txs in block if exist
		txs, total, err = s.cacheClient.TxsByBlockHeight(ctx, height, pagination)
		if err != nil {
			s.logger.Debug("cannot get block txs by height from cache", zap.String("blockHeight", block), zap.Error(err))
			// otherwise, get from db
			txs, total, err = s.dbClient.TxsByBlockHeight(ctx, height, pagination)
			if err != nil {
				s.logger.Warn("cannot get block txs by height from db", zap.String("blockHeight", block), zap.Error(err))
				// try to get block txs from RPC
				blockRPC, err := s.kaiClient.BlockByHeight(ctx, height)
				if err != nil {
					s.logger.Warn("cannot get block txs by height from RPC", zap.String("blockHeight", block), zap.Error(err))
					return api.InternalServer.Build(c)
				}
				txs = blockRPC.Txs
				if pagination.Skip > len(txs) {
					txs = []*types.Transaction(nil)
				} else if pagination.Skip+pagination.Limit > len(txs) {
					txs = blockRPC.Txs[pagination.Skip:len(txs)]
				} else {
					txs = blockRPC.Txs[pagination.Skip : pagination.Skip+pagination.Limit]
				}
				total = blockRPC.NumTxs
				s.Logger.Debug("got block txs by height from RPC:", zap.String("blockHeight", block))
			} else {
				s.Logger.Debug("got block txs by height from db:", zap.String("blockHeight", block))
			}
		} else {
			s.Logger.Debug("got block txs by height from cache:", zap.String("blockHeight", block))
		}
	}

	smcAddress := s.getValidatorsAddressAndRole(ctx)
	var result Transactions
	for _, tx := range txs {
		t := SimpleTransaction{
			Hash:             tx.Hash,
			BlockNumber:      tx.BlockNumber,
			Time:             tx.Time,
			From:             tx.From,
			To:               tx.To,
			ContractAddress:  tx.ContractAddress,
			Value:            tx.Value,
			TxFee:            tx.TxFee,
			Status:           tx.Status,
			DecodedInputData: tx.DecodedInputData,
		}
		if smcAddress[tx.To] != nil {
			t.ToName = smcAddress[tx.To].Name
			t.Role = smcAddress[tx.To].Role
			t.IsInValidatorsList = true
		}
		if tx.To == cfg.StakingContractAddr {
			t.ToName = cfg.StakingContractName
		}
		result = append(result, t)
	}

	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  result,
	}).Build(c)
}

func (s *Server) BlocksByProposer(c echo.Context) error {
	ctx := context.Background()
	pagination, page, limit := getPagingOption(c)
	blocks, total, err := s.dbClient.BlocksByProposer(ctx, c.Param("address"), pagination)
	if err != nil {
		s.logger.Debug("Cannot get blocks by proposer from db", zap.Error(err))
		return api.Invalid.Build(c)
	}
	smcAddress := s.getValidatorsAddressAndRole(ctx)
	var result Blocks
	for _, block := range blocks {
		b := SimpleBlock{
			Height:          block.Height,
			Hash:            block.Hash,
			Time:            block.Time,
			ProposerAddress: block.ProposerAddress,
			NumTxs:          block.NumTxs,
			GasLimit:        block.GasLimit,
			GasUsed:         block.GasUsed,
			Rewards:         block.Rewards,
		}
		b.ProposerName = smcAddress[b.ProposerAddress].Name
		result = append(result, b)
	}
	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  result,
	}).Build(c)
}

func (s *Server) Txs(c echo.Context) error {
	ctx := context.Background()
	pagination, page, limit := getPagingOption(c)
	var (
		err error
		txs []*types.Transaction
	)

	txs, err = s.cacheClient.LatestTransactions(ctx, pagination)
	if err != nil || txs == nil || len(txs) < limit {
		s.logger.Debug("Cannot get latest txs from cache", zap.Error(err))
		txs, err = s.dbClient.LatestTxs(ctx, pagination)
		if err != nil {
			s.logger.Debug("Cannot get latest txs from db", zap.Error(err))
			return api.Invalid.Build(c)
		}
		s.logger.Debug("Got latest txs from db")
	} else {
		s.logger.Debug("Got latest txs from cached")
	}

	smcAddress := s.getValidatorsAddressAndRole(ctx)
	var result Transactions
	for _, tx := range txs {
		t := SimpleTransaction{
			Hash:             tx.Hash,
			BlockNumber:      tx.BlockNumber,
			Time:             tx.Time,
			From:             tx.From,
			To:               tx.To,
			ContractAddress:  tx.ContractAddress,
			Value:            tx.Value,
			TxFee:            tx.TxFee,
			Status:           tx.Status,
			DecodedInputData: tx.DecodedInputData,
		}

		if smcAddress[tx.To] != nil {
			t.ToName = smcAddress[tx.To].Name
			t.Role = smcAddress[tx.To].Role
			t.IsInValidatorsList = true
		}

		if tx.To == cfg.StakingContractAddr {
			t.ToName = cfg.StakingContractName
		}
		result = append(result, t)
	}

	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: s.cacheClient.TotalTxs(ctx),
		Data:  result,
	}).Build(c)
}

func (s *Server) Addresses(c echo.Context) error {
	ctx := context.Background()
	pagination, page, limit := getPagingOption(c)
	sortDirectionStr := c.QueryParam("sort")
	sortDirection, err := strconv.Atoi(sortDirectionStr)
	if err != nil || (sortDirection != 1 && sortDirection != -1) {
		sortDirection = -1 // DESC
	}
	addrs, err := s.dbClient.GetListAddresses(ctx, sortDirection, pagination)
	if err != nil {
		return api.Invalid.Build(c)
	}
	totalHolders, totalContracts := s.cacheClient.TotalHolders(ctx)
	smcAddress := s.getValidatorsAddressAndRole(ctx)
	var result Addresses
	for _, addr := range addrs {
		addrInfo := SimpleAddress{
			Address:       addr.Address,
			BalanceString: addr.BalanceString,
			IsContract:    addr.IsContract,
			Name:          addr.Name,
		}
		if smcAddress[addr.Address] != nil {
			addrInfo.IsInValidatorsList = true
			addrInfo.Role = smcAddress[addr.Address].Role
			addrInfo.Name = smcAddress[addr.Address].Name
		}
		result = append(result, addrInfo)
	}
	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: totalHolders + totalContracts,
		Data:  result,
	}).Build(c)
}

func (s *Server) AddressInfo(c echo.Context) error {
	ctx := context.Background()
	address := c.Param("address")
	smcAddress := s.getValidatorsAddressAndRole(ctx)
	addrInfo, err := s.dbClient.AddressByHash(ctx, address)
	if err == nil {
		result := SimpleAddress{
			Address:       addrInfo.Address,
			BalanceString: addrInfo.BalanceString,
			IsContract:    addrInfo.IsContract,
			Name:          addrInfo.Name,
		}
		if smcAddress[result.Address] != nil {
			result.IsInValidatorsList = true
			result.Role = smcAddress[result.Address].Role
			result.Name = smcAddress[result.Address].Name
		}
		return api.OK.SetData(result).Build(c)
	}
	s.logger.Warn("address not found in db, getting from RPC instead...", zap.Error(err))
	// try to get balance and code at this address to determine whether we should write this address info to database or not
	balance, err := s.kaiClient.GetBalance(ctx, address)
	if err != nil {
		return err
	}
	balanceInBigInt, _ := new(big.Int).SetString(balance, 10)
	balanceFloat, _ := new(big.Float).SetPrec(100).Quo(new(big.Float).SetInt(balanceInBigInt), new(big.Float).SetInt(cfg.Hydro)).Float64() //converting to KAI from HYDRO
	addrInfo = &types.Address{
		Address:       address,
		BalanceFloat:  balanceFloat,
		BalanceString: balance,
		IsContract:    false,
	}
	code, err := s.kaiClient.GetCode(ctx, address)
	if err == nil && len(code) > 0 {
		addrInfo.IsContract = true
	}
	// write this address to db if its balance is larger than 0 or it's a SMC
	if balance != "0" || addrInfo.IsContract {
		// update name of this address
		vals, err := s.cacheClient.Validators(ctx)
		if err != nil {
			vals = &types.Validators{
				Validators: []*types.Validator{},
			}
		}
		cmnAddress := common.HexToAddress(address)
		for _, val := range vals.Validators {
			if val.SmcAddress.Equal(cmnAddress) {
				addrInfo.Name = val.Name
				break
			}
		}
		_ = s.dbClient.InsertAddress(ctx, addrInfo) // insert this address to database
	}
	return api.OK.SetData(SimpleAddress{
		Address:       addrInfo.Address,
		BalanceString: addrInfo.BalanceString,
		IsContract:    addrInfo.IsContract,
		Name:          addrInfo.Name,
	}).Build(c)
}

func (s *Server) AddressTxs(c echo.Context) error {
	ctx := context.Background()
	var err error
	address := c.Param("address")
	pagination, page, limit := getPagingOption(c)

	txs, total, err := s.dbClient.TxsByAddress(ctx, address, pagination)
	if err != nil {
		s.logger.Debug("error while get address txs:", zap.Error(err))
		return err
	}

	smcAddress := s.getValidatorsAddressAndRole(ctx)
	var result Transactions
	for _, tx := range txs {
		t := SimpleTransaction{
			Hash:             tx.Hash,
			BlockNumber:      tx.BlockNumber,
			Time:             tx.Time,
			From:             tx.From,
			To:               tx.To,
			ContractAddress:  tx.ContractAddress,
			Value:            tx.Value,
			TxFee:            tx.TxFee,
			Status:           tx.Status,
			DecodedInputData: tx.DecodedInputData,
		}
		if smcAddress[tx.To] != nil {
			t.ToName = smcAddress[tx.To].Name
			t.Role = smcAddress[tx.To].Role
			t.IsInValidatorsList = true
		}
		if tx.To == cfg.StakingContractAddr {
			t.ToName = cfg.StakingContractName
		}
		result = append(result, t)
	}

	s.logger.Info("address txs:", zap.String("address", address))
	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  result,
	}).Build(c)
}

func (s *Server) AddressHolders(c echo.Context) error {
	var (
		page, limit int
		err         error
	)
	//blockHash := c.Param("blockHash")
	pageParams := c.QueryParam("page")
	limitParams := c.QueryParam("limit")
	page, err = strconv.Atoi(pageParams)
	if err != nil {
		page = 0
	}
	limit, err = strconv.Atoi(limitParams)
	if err != nil {
		limit = 20
	}

	var holders []*types.TokenHolder
	for i := 0; i < limit; i++ {
		holder := &types.TokenHolder{}
		if err := faker.FakeData(&holder); err != nil {
			return err
		}
		holders = append(holders, holder)
	}

	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: uint64(limit * 15),
		Data:  holders,
	}).Build(c)
}

func (s *Server) AddressOwnedTokens(c echo.Context) error {
	return api.OK.Build(c)
}

// AddressInternal
func (s *Server) AddressInternalTxs(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) AddressContract(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) AddressTxByNonce(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) AddressTxHashByNonce(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) TxByHash(c echo.Context) error {
	ctx := context.Background()
	txHash := c.Param("txHash")
	if txHash == "" {
		return api.Invalid.Build(c)
	}

	var tx *types.Transaction
	tx, err := s.dbClient.TxByHash(ctx, txHash)
	if err != nil {
		// try to get tx by hash through RPC
		s.Logger.Debug("cannot get tx by hash from db:", zap.String("txHash", txHash))
		tx, err = s.kaiClient.GetTransaction(ctx, txHash)
		if err != nil {
			s.Logger.Warn("cannot get tx by hash from RPC:", zap.String("txHash", txHash))
			return api.Invalid.Build(c)
		}
		receipt, err := s.kaiClient.GetTransactionReceipt(ctx, txHash)
		if err != nil {
			s.Logger.Warn("cannot get receipt by hash from RPC:", zap.String("txHash", txHash))
		}
		s.Logger.Debug("got tx by hash from RPC:", zap.String("txHash", txHash))
		if receipt != nil {
			tx.Logs = receipt.Logs
			tx.Root = receipt.Root
			tx.Status = receipt.Status
			tx.GasUsed = receipt.GasUsed
			tx.ContractAddress = receipt.ContractAddress
		}
	} else {
		s.Logger.Debug("got tx by hash from db:", zap.String("txHash", txHash))
	}

	// add name of recipient, if any
	if decoded, err := s.kaiClient.DecodeInputData(tx.To, tx.InputData); err == nil {
		tx.DecodedInputData = decoded
	}
	result := &Transaction{
		BlockHash:        tx.BlockHash,
		BlockNumber:      tx.BlockNumber,
		Hash:             tx.Hash,
		From:             tx.From,
		To:               tx.To,
		Status:           tx.Status,
		ContractAddress:  tx.ContractAddress,
		Value:            tx.Value,
		GasPrice:         tx.GasPrice,
		GasLimit:         tx.GasLimit,
		GasUsed:          tx.GasUsed,
		TxFee:            tx.TxFee,
		Nonce:            tx.Nonce,
		Time:             tx.Time,
		InputData:        tx.InputData,
		DecodedInputData: tx.DecodedInputData,
		Logs:             tx.Logs,
		TransactionIndex: tx.TransactionIndex,
		LogsBloom:        tx.LogsBloom,
		Root:             tx.Root,
	}
	if result.To == cfg.StakingContractAddr {
		result.ToName = cfg.StakingContractName
		return api.OK.SetData(result).Build(c)
	}
	smcAddress := s.getValidatorsAddressAndRole(ctx)
	if smcAddress[result.To] != nil {
		result.ToName = smcAddress[result.To].Name
		result.Role = smcAddress[result.To].Role
		result.IsInValidatorsList = true
		return api.OK.SetData(result).Build(c)
	}

	return api.OK.SetData(result).Build(c)
}

func (s *Server) TxExist(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) Contracts(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) BlockTime(c echo.Context) error {
	panic("implement me")
}

func (s *Server) getValidatorsList(ctx context.Context) (*types.Validators, error) {
	valsList, err := s.kaiClient.Validators(ctx)
	if err != nil {
		s.logger.Warn("cannot get validators list from RPC", zap.Error(err))
		return nil, err
	}
	s.logger.Debug("Got validators list from RPC")
	err = s.cacheClient.UpdateValidators(ctx, valsList)
	if err != nil {
		s.logger.Warn("cannot store validators list to cache", zap.Error(err))
	}
	return valsList, nil
}

func getPagingOption(c echo.Context) (*types.Pagination, int, int) {
	pageParams := c.QueryParam("page")
	limitParams := c.QueryParam("limit")
	page, err := strconv.Atoi(pageParams)
	if err != nil {
		page = 0
	}
	limit, err := strconv.Atoi(limitParams)
	if err != nil {
		limit = 10
	}
	pagination := &types.Pagination{
		Skip:  page * limit,
		Limit: limit,
	}
	pagination.Sanitize()
	return pagination, page, limit
}

func (s *Server) getValidatorsAddressAndRole(ctx context.Context) map[string]*valInfoResponse {
	vals, err := s.kaiClient.Validators(ctx)
	if err != nil {
		vals, err = s.getValidatorsList(ctx)
		if err != nil {
			vals = nil
		}
	}

	smcAddress := map[string]*valInfoResponse{}
	for _, v := range vals.Validators {
		smcAddress[v.SmcAddress.String()] = &valInfoResponse{
			Name: v.Name,
			Role: v.Role,
		}
		smcAddress[v.Address.String()] = &valInfoResponse{
			Name: v.Name,
			Role: v.Role,
		}
	}
	return smcAddress
}
