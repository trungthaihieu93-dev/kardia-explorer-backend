// Package cache
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/types"
)

const (
	KeyLatestBlockHeight = "#block#latestHeight"
	KeyLatestBlock       = "#block#latest"

	KeyBlocks        = "#blocks" // List
	KeyBlockByNumber = "#block#%d"
	KeyBlockByHash   = "#block#%s"

	KeyTxsOfBlockIndex       = "#block#index#%d#txs"
	KeyTxOfBlockIndexByNonce = "#block#index#%d#tx#%d"
	KeyTxOfBlockIndexByHash  = "#block#index#%d#tx#%s"

	KeyLatestStats = "#stats#latest"

	PatternGetAllKeyOfBlockIndex = "#block#index#%d*"

	ErrorBlocks = "#errorBlocks" // List
)

type Redis struct {
	cfg    Config
	client *redis.Client

	logger *zap.Logger
}

func (c *Redis) PopReceipt(ctx context.Context) (*types.Receipt, error) {
	panic("implement me")
}

func (c *Redis) BlocksSize(ctx context.Context) (int64, error) {
	size, err := c.client.LLen(ctx, KeyBlocks).Result()
	if err != nil {
		return -1, err
	}
	if size == 0 {
		return 0, errors.New("blocks has no cache")
	}

	return size, nil
}

func (c *Redis) InsertTxs(ctx context.Context, txs []*types.Transaction) error {
	if len(txs) == 0 {
		return nil
	}
	// Get block index
	var blockIndex int
	if err := c.client.Get(ctx, fmt.Sprintf(KeyBlockByNumber, txs[0].BlockNumber)).Scan(&blockIndex); err != nil {
		c.logger.Debug("cannot get block at index", zap.Int("BlockIndex", blockIndex))
		return err
	}
	// todo: benchmark with different size of txs
	// todo: this way look quite stupid
	for _, tx := range txs {
		txStr, err := json.Marshal(tx)
		if err != nil {
			c.logger.Debug("cannot marshal txs arr to string")
			return err
		}

		txIndex, err := c.client.LPush(ctx, KeyTxsOfBlockIndex, txStr).Result()
		if err != nil {
			c.logger.Debug("cannot insert txs")
			return err
		}
		txOfBlockIndexByNonce := fmt.Sprintf(KeyTxOfBlockIndexByNonce, blockIndex, tx.Nonce)
		txOfBlockIndexByHash := fmt.Sprintf(KeyTxOfBlockIndexByHash, blockIndex, tx.Hash)
		if err := c.client.Set(ctx, txOfBlockIndexByNonce, txIndex, c.cfg.DefaultExpiredTime).Err(); err != nil {
			return err
		}
		if err := c.client.Set(ctx, txOfBlockIndexByHash, txIndex, c.cfg.DefaultExpiredTime).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Redis) TxByHash(ctx context.Context, txHash string) (*types.Transaction, error) {
	panic("implement me")
}

// ImportBlock cache follow step:
// Keep recent N blocks in memory as cache and temp write DB
// Maintain SetByHash and SetByNumber return blockIndex
func (c *Redis) InsertBlock(ctx context.Context, block *types.Block) error {
	c.logger.Debug("Start insert block")
	size, err := c.client.LLen(ctx, KeyBlocks).Result()
	if err != nil {
		c.logger.Debug("cannot get size of #blocks", zap.Error(err))
		return err
	}

	// Size over buffer then
	c.logger.Debug("redis block buffer size: ", zap.Int64("size", size), zap.Int64("c.cfg.BlockBuffer", c.cfg.BlockBuffer))
	if size >= c.cfg.BlockBuffer && size != 0 {
		// Delete block at last index
		if err := c.deleteKeysOfBlockIndex(ctx, size); err != nil {
			c.logger.Debug("cannot delete keys of block index", zap.Error(err))
			return err
		}

		if _, err := c.client.RPop(ctx, KeyBlocks).Result(); err != nil {
			c.logger.Debug("cannot pop last element", zap.Error(err))
			return err
		}
	}

	// Push new block to list
	// Using clone instead assign
	blockCache := *block
	blockCache.Txs = []*types.Transaction{}
	blockCache.Receipts = []*types.Receipt{}

	lPushResult := c.client.LPush(ctx, KeyBlocks, blockCache.String())
	blockIndex, err := lPushResult.Result()
	if err != nil {
		return err
	}

	c.logger.Debug("Push new block success", zap.Int64("index", blockIndex))
	blockByHashKey := fmt.Sprintf(KeyBlockByHash, block.Hash)
	blockByNumberKey := fmt.Sprintf(KeyBlockByNumber, block.Height)
	if err := c.client.Set(ctx, blockByHashKey, blockIndex, 0).Err(); err != nil {
		return err
	}
	if err := c.client.Set(ctx, blockByNumberKey, blockIndex, 0).Err(); err != nil {
		return err
	}
	return nil
}

func (c *Redis) BlockByHash(ctx context.Context, blockHash string) (*types.Block, error) {
	key := fmt.Sprintf(KeyBlockByHash, blockHash)
	var index int64
	if err := c.client.Get(ctx, key).Scan(&index); err != nil {
		return nil, err
	}
	return c.getBlockIndex(ctx, index)
}

func (c *Redis) BlockByHeight(ctx context.Context, blockHeight uint64) (*types.Block, error) {
	key := fmt.Sprintf(KeyBlockByNumber, blockHeight)
	var index int64
	if err := c.client.Get(ctx, key).Scan(&index); err != nil {
		return nil, err
	}
	return c.getBlockIndex(ctx, index)
}

func (c *Redis) LatestBlocks(ctx context.Context, pagination *types.Pagination) ([]*types.Block, error) {
	var (
		blockList        []*types.Block
		marshalledBlocks []string
		startIndex       int64 = 0 + int64(pagination.Skip)
		endIndex         int64 = startIndex + int64(pagination.Limit) - 1
	)
	c.logger.Debug("Getting blocks from cache: ", zap.Int("pagination.Skip", pagination.Skip), zap.Int("pagination.Limit", pagination.Limit))
	marshalledBlocks, err := c.client.LRange(ctx, KeyBlocks, startIndex, endIndex).Result()
	c.logger.Debug("Getting blocks from cache: ", zap.Int64("startIndex", startIndex), zap.Int64("endIndex", endIndex))
	c.logger.Debug("Blocks from cache: ", zap.Any("marshalledBlocks", marshalledBlocks))
	if err != nil {
		return nil, err
	}
	for _, bStr := range marshalledBlocks {
		var b types.Block
		err := json.Unmarshal([]byte(bStr), &b)
		if err != nil {
			return nil, err
		}
		blockList = append(blockList, &b)
	}
	c.logger.Debug("Blocks from cache: ", zap.Any("blocks", blockList))
	return blockList, nil
}

func (c *Redis) LatestTransactions(ctx context.Context, pagination *types.Pagination) ([]*types.Transaction, error) {
	var (
		txList        []*types.Transaction
		marshalledTxs []string
		startIndex    int64 = 0 - int64(pagination.Skip)
		endIndex      int64 = startIndex + int64(pagination.Limit) - 1
	)
	if endIndex > 0 {
		endIndex = 0
	}
	marshalledTxs, err := c.client.LRange(ctx, KeyBlocks, startIndex, endIndex).Result()
	c.logger.Debug("Getting Txs from cache: ", zap.Int64("startIndex", startIndex), zap.Int64("endIndex", endIndex))
	if err != nil {
		return nil, err
	}
	for _, txStr := range marshalledTxs {
		var tx types.Transaction
		err := json.Unmarshal([]byte(txStr), &tx)
		if err != nil {
			return nil, err
		}
		txList = append(txList, &tx)
	}
	c.logger.Debug("Txs from cache: ", zap.Any("blocks", txList))
	return txList, nil
}

func (c *Redis) InsertErrorBlocks(ctx context.Context, start uint64, end uint64) error {
	for i := start + 1; i < end; i++ {
		_, err := c.client.LPush(ctx, ErrorBlocks, strconv.FormatUint(i, 10)).Result()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Redis) PopErrorBlockHeight(ctx context.Context) (uint64, error) {
	heightStr, err := c.client.LPop(ctx, ErrorBlocks).Result()
	if err != nil {
		return 0, err
	}
	height, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return height, nil
}

func (c *Redis) getBlockIndex(ctx context.Context, index int64) (*types.Block, error) {
	block := &types.Block{}
	lIndexResult := c.client.LIndex(ctx, KeyBlocks, index)
	if err := lIndexResult.Err(); err != nil {
		return nil, err
	}

	if err := lIndexResult.Scan(block); err != nil {
		return nil, err
	}
	return nil, nil
}

func (c *Redis) deleteKeysOfBlockIndex(ctx context.Context, blockIndex int64) error {
	pattern := fmt.Sprintf(PatternGetAllKeyOfBlockIndex, blockIndex)
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		c.logger.Debug("cannot get keys", zap.Error(err))
		return err
	}

	var blockStr string
	if err := c.client.LIndex(ctx, KeyBlocks, blockIndex-1).Scan(&blockStr); err != nil {
		c.logger.Debug("cannot get block", zap.Int64("BlockIndex", blockIndex-1), zap.Error(err))
		return err
	}
	var block types.Block
	if err := json.Unmarshal([]byte(blockStr), &block); err != nil {
		c.logger.Debug("cannot marshal block from cache to object", zap.Error(err))
		return err
	}

	c.logger.Debug("Block info", zap.Any("block", block))

	keys = append(keys, []string{
		fmt.Sprintf(KeyBlockByNumber, block.Height),
		fmt.Sprintf(KeyBlockByHash, block.Hash),
	}...)

	if _, err := c.client.Del(ctx, keys...).Result(); err != nil {
		c.logger.Debug("cannot delete keys", zap.Strings("Keys", keys))
		return err
	}

	return nil
}
