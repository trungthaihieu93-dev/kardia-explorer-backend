// Package v1
package v1

import (
	"github.com/labstack/echo"
)

// EchoServer define all API expose
type EchoServer interface {
	// General
	Ping(c echo.Context) error
	Stats(c echo.Context) error
	TotalHolders(c echo.Context) error
	TokenInfo(c echo.Context) error
	Nodes(c echo.Context) error

	// Staking-related
	StakingStats(c echo.Context) error
	Validator(c echo.Context) error
	ValidatorsByDelegator(c echo.Context) error
	Validators(c echo.Context) error
	Candidates(c echo.Context) error
	MobileValidators(c echo.Context) error
	MobileCandidates(c echo.Context) error

	// Proposal
	GetProposalsList(c echo.Context) error
	GetProposalDetails(c echo.Context) error
	GetParams(c echo.Context) error

	// Blocks
	Blocks(c echo.Context) error
	Block(c echo.Context) error
	BlockTxs(c echo.Context) error
	BlocksByProposer(c echo.Context) error
	PersistentErrorBlocks(c echo.Context) error

	// Addresses
	Addresses(c echo.Context) error
	AddressInfo(c echo.Context) error
	AddressTxs(c echo.Context) error
	AddressHolders(c echo.Context) error

	// Tx
	Txs(c echo.Context) error
	TxByHash(c echo.Context) error

	// Admin sector
	ReloadAddressesBalance(c echo.Context) error
	ReloadValidators(c echo.Context) error
	UpdateAddressName(c echo.Context) error
	UpsertNetworkNodes(c echo.Context) error
	RemoveNetworkNodes(c echo.Context) error
	UpdateSupplyAmounts(c echo.Context) error

	IContract

	//
	SearchAddressByName(c echo.Context) error

	GetHoldersListByToken(c echo.Context) error
	GetInternalTxs(c echo.Context) error
}

type IContract interface {
	Contracts(c echo.Context) error
	Contract(c echo.Context) error
	InsertContract(c echo.Context) error
	UpdateContract(c echo.Context) error
	UpdateSMCABIByType(c echo.Context) error

	ContractEvents(c echo.Context) error
}

type restDefinition struct {
	method      string
	path        string
	fn          func(c echo.Context) error
	middlewares []echo.MiddlewareFunc
}

func BindAPI(gr *echo.Group, srv EchoServer) error {

	v1Gr := gr.Group("/api/v1")
	apis := []restDefinition{
		{
			method:      echo.GET,
			path:        "/ping",
			fn:          srv.Ping,
			middlewares: nil,
		},
		{
			method: echo.GET,
			path:   "/dashboard/stats",
			fn:     srv.Stats,
		},
		{
			method: echo.GET,
			path:   "/dashboard/holders/total",
			fn:     srv.TotalHolders,
		},
		{
			method: echo.GET,
			path:   "/dashboard/token",
			fn:     srv.TokenInfo,
		},
		{
			method: echo.PUT,
			path:   "/dashboard/token/supplies",
			fn:     srv.UpdateSupplyAmounts,
		},
		{
			method: echo.PUT,
			path:   "/nodes",
			fn:     srv.UpsertNetworkNodes,
		},
		{
			method: echo.DELETE,
			path:   "/nodes/:nodeID",
			fn:     srv.RemoveNetworkNodes,
		},
		// Blocks
		{
			method: echo.GET,
			// Query params: ?page=0&limit=10
			path: "/blocks",
			fn:   srv.Blocks,
		},
		{
			method: echo.GET,
			path:   "/blocks/:block",
			fn:     srv.Block,
		},
		{
			method: echo.GET,
			path:   "/blocks/error",
			fn:     srv.PersistentErrorBlocks,
		},
		{
			method: echo.GET,
			// Params: proposer address
			// Query params: ?page=0&limit=10
			path:        "/blocks/proposer/:address",
			fn:          srv.BlocksByProposer,
			middlewares: nil,
		},
		{
			method: echo.GET,
			// Params: block's hash
			// Query params: ?page=0&limit=10
			path:        "/block/:block/txs",
			fn:          srv.BlockTxs,
			middlewares: nil,
		},
		{
			method: echo.GET,
			path:   "/txs/:txHash",
			fn:     srv.TxByHash,
		},
		{
			method: echo.GET,
			// Query params: ?page=0&limit=10
			path:        "/txs",
			fn:          srv.Txs,
			middlewares: nil,
		},
		// Address
		{
			method: echo.GET,
			// Query params: ?page=0&limit=10&sort=1
			path: "/addresses",
			fn:   srv.Addresses,
		},
		{
			method: echo.GET,
			path:   "/addresses/:address",
			fn:     srv.AddressInfo,
		},
		{
			method: echo.POST,
			path:   "/addresses/reload",
			fn:     srv.ReloadAddressesBalance,
		},
		// Tokens
		{
			method:      echo.GET,
			path:        "/addresses/:address/txs",
			fn:          srv.AddressTxs,
			middlewares: nil,
		},
		{
			method: echo.GET,
			// Query params: ?page=0&limit=10&contractAddress=0x
			path:        "/addresses/:address/tokens",
			fn:          srv.AddressHolders,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/nodes",
			fn:          srv.Nodes,
			middlewares: nil,
		},
		// Proposal
		{
			method: echo.GET,
			// Query params: ?page=0&limit=10
			path:        "/proposal",
			fn:          srv.GetProposalsList,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/proposal/:id",
			fn:          srv.GetProposalDetails,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/proposal/params",
			fn:          srv.GetParams,
			middlewares: nil,
		},
		{
			method:      echo.PUT,
			path:        "/addresses",
			fn:          srv.UpdateAddressName,
			middlewares: nil,
		},
		{
			method:      echo.POST,
			path:        "/validators/reload",
			fn:          srv.ReloadValidators,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/search",
			fn:          srv.SearchAddressByName,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/token/holders/:contractAddress",
			fn:          srv.GetHoldersListByToken,
			middlewares: nil,
		},
		{
			method: echo.GET,
			// Query params: ?page=0&limit=10&address=0x&contractAddress=0x
			path:        "/token/txs",
			fn:          srv.GetInternalTxs,
			middlewares: nil,
		},
	}
	bindContractAPIs(v1Gr, srv)
	bindStakingAPIs(v1Gr, srv)
	for _, api := range apis {
		v1Gr.Add(api.method, api.path, api.fn, api.middlewares...)
	}

	return nil
}

func bindContractAPIs(gr *echo.Group, srv EchoServer) {
	apis := []restDefinition{
		{
			method:      echo.POST,
			path:        "/contracts",
			fn:          srv.InsertContract,
			middlewares: nil,
		},
		{
			method:      echo.PUT,
			path:        "/contracts",
			fn:          srv.UpdateContract,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/contracts",
			fn:          srv.Contracts,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/contracts/:contractAddress",
			fn:          srv.Contract,
			middlewares: nil,
		},
		{
			method:      echo.PUT,
			path:        "/contracts/abi",
			fn:          srv.UpdateSMCABIByType,
			middlewares: nil,
		},
		{
			method: echo.GET,
			// Query params: ?page=0&limit=10&contractAddress=0x&methodName=0x&txHash=0x
			path:        "/contracts/events",
			fn:          srv.ContractEvents,
			middlewares: nil,
		},
	}
	for _, api := range apis {
		gr.Add(api.method, api.path, api.fn, api.middlewares...)
	}

}

func bindStakingAPIs(gr *echo.Group, srv EchoServer) {
	apis := []restDefinition{
		//Validator
		{
			method:      echo.GET,
			path:        "/staking/stats",
			fn:          srv.StakingStats,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/validators/:address",
			fn:          srv.Validator,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/delegators/:address/validators",
			fn:          srv.ValidatorsByDelegator,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/validators/candidates",
			fn:          srv.MobileCandidates,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/validators",
			fn:          srv.MobileValidators,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/staking/candidates",
			fn:          srv.Candidates,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/staking/validators",
			fn:          srv.Validators,
			middlewares: nil,
		},
	}
	for _, api := range apis {
		gr.Add(api.method, api.path, api.fn, api.middlewares...)
	}
}
