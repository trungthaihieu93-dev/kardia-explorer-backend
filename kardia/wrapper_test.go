// Package kardia
package kardia

import (
	"context"
	"fmt"
	"testing"

	"github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/go-kardia/types/time"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/types"
)

func TestWrapper_Validator(t *testing.T) {
	ctx := context.Background()
	logger, err := zap.NewDevelopment()
	assert.Nil(t, err)
	node1, err := kardia.NewNode("https://kai-internal-1.kardiachain.io", logger)
	node2, err := kardia.NewNode("https://kai-internal-1.kardiachain.io", logger)
	node3, err := kardia.NewNode("https://kai-internal-1.kardiachain.io", logger)
	w := &Wrapper{
		trustedNodes: []kardia.Node{node1, node2, node3},
		logger:       logger,
	}
	vSMC := "0x35a501fA8368c9f6D6B4EA5D7daDc963091B6F97"

	v, err := w.Validator(ctx, vSMC)
	assert.Nil(t, err)

	fmt.Println("Commission", v.CommissionRate)

	valSets, err := w.TrustedNode().ValidatorSets(ctx)
	assert.Nil(t, err)
	fmt.Println("ValSet", len(valSets))

	if v.StakedAmount == "0" {
		fmt.Println("Should removed ")
	}

	fmt.Printf("Validator : %+v \n", v)
}

func TestWrapper_Validators(t *testing.T) {
	ctx := context.Background()
	logger, err := zap.NewDevelopment()
	assert.Nil(t, err)
	node, err := kardia.NewNode("https://kai-internal-1.kardiachain.io", logger)
	node1, err := kardia.NewNode("https://kai-internal-1.kardiachain.io", logger)
	assert.Nil(t, err)
	node2, err := kardia.NewNode("https://kai-internal-1.kardiachain.io", logger)
	assert.Nil(t, err)
	node3, err := kardia.NewNode("https://kai-internal-1.kardiachain.io", logger)
	assert.Nil(t, err)
	//node4, err := kardia.NewNode("https://kai-ecosystem-4.kardiachain.io", logger)
	//assert.Nil(t, err)
	//node5, err := kardia.NewNode("https://kai-ecosystem-5.kardiachain.io", logger)
	//assert.Nil(t, err)
	//node6, err := kardia.NewNode("https://kai-ecosystem-6.kardiachain.io", logger)
	//assert.Nil(t, err)
	w := &Wrapper{
		trustedNodes: []kardia.Node{node},
		publicNodes:  []kardia.Node{node, node1, node2, node3},
		logger:       logger,
	}
	var validators []*types.Validator

	startTimeNoWorker := time.Now()
	validators, err = w.Validators(ctx)
	assert.Nil(t, err)
	fmt.Println("TotalTime no worker", time.Now().Sub(startTimeNoWorker))
	startTime := time.Now()
	validators, err = w.ValidatorsWithWorker(ctx)
	assert.Nil(t, err)
	fmt.Println("TotalTime with worker", time.Now().Sub(startTime))
	dbCfg := db.Config{
		DbAdapter: "mgo",
		DbName:    "testDB",
		URL:       "mongodb://kardia.ddns.net:27017",
		MinConn:   1,
		MaxConn:   4,
		FlushDB:   false,
		Logger:    logger,
	}
	dbClient, err := db.NewClient(dbCfg)
	assert.Nil(t, err)
	assert.Nil(t, dbClient.UpsertValidators(ctx, validators))
}

func TestWrapper_Delegators(t *testing.T) {
	ctx := context.Background()
	logger, err := zap.NewDevelopment()
	assert.Nil(t, err)
	node, err := kardia.NewNode("https://rpc.kardiachain.io", logger)
	node1, err := kardia.NewNode("https://kai-ecosystem-1.kardiachain.io", logger)
	assert.Nil(t, err)
	node2, err := kardia.NewNode("https://kai-ecosystem-2.kardiachain.io", logger)
	assert.Nil(t, err)
	node3, err := kardia.NewNode("https://kai-ecosystem-3.kardiachain.io", logger)
	assert.Nil(t, err)
	//node4, err := kardia.NewNode("https://kai-ecosystem-4.kardiachain.io", logger)
	//assert.Nil(t, err)
	//node5, err := kardia.NewNode("https://kai-ecosystem-5.kardiachain.io", logger)
	//assert.Nil(t, err)
	//node6, err := kardia.NewNode("https://kai-ecosystem-6.kardiachain.io", logger)
	//assert.Nil(t, err)
	w := &Wrapper{
		trustedNodes: []kardia.Node{node},
		publicNodes:  []kardia.Node{node, node1, node2, node3},
	}
	var delegators []*types.Delegator
	validatorSMCAddr := "0x4dAe614b2eA2FaeeDDE7830A2e7fcEDdAE9f9161"
	startTimeNoWorker := time.Now()
	delegators, err = w.Delegators(ctx, validatorSMCAddr)
	assert.Nil(t, err)
	fmt.Println("TotalTime no worker", time.Now().Sub(startTimeNoWorker))
	startTime := time.Now()
	delegators, err = w.DelegatorsWithWorker(ctx, validatorSMCAddr)
	assert.Nil(t, err)
	fmt.Println("TotalTime with worker", time.Now().Sub(startTime))

	fmt.Println("Size ", len(delegators))

	//for _, v := range validators {
	//	fmt.Printf("Validator: %+v \n", v)
	//}
	//dbCfg := db.Config{
	//	DbAdapter: "mgo",
	//	DbName:    "testDB",
	//	URL:       "mongodb://kardia.ddns.net:27017",
	//	MinConn:   1,
	//	MaxConn:   4,
	//	FlushDB:   false,
	//	Logger:    logger,
	//}
	//dbClient, err := db.NewClient(dbCfg)
	//assert.Nil(t, err)
	//assert.Nil(t, dbClient.UpsertDe(ctx, validators))
}
