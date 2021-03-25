// Package types
package types

type ContractType string

const (
	ContractTypeDefault   ContractType = "Normal"
	ContractTypeStaking                = "Staking"
	ContractTypeParams                 = "Params"
	ContractTypeValidator              = "Validator"
	ContractTypeKRC20                  = "KRC20"
	ContractTypeKRC721                 = "KRC721"
)

type ContractStatus int

const (
	ContractStatusUnknown ContractStatus = iota
	ContractStatusUnverified
	ContractStatusVerified
)

// Contract define simple information about a SMC in kardia system
type Contract struct {
	Address      string `json:"address" bson:"address"`
	Name         string `json:"name" bson:"name"`
	ABI          string `json:"abi" bson:"abi"`
	Bytecode     string `json:"bytecode,omitempty" bson:"bytecode"`
	OwnerAddress string `json:"ownerAddress,omitempty" bson:"ownerAddress"`
	// Trace
	TxHash        string       `json:"txHash,omitempty" bson:"txHash"`
	CreateAtBlock uint64       `json:"-" bson:"createAtBlock"`
	Type          ContractType `json:"type" bson:"type"`
	Description   string       `json:"description" bson:"description"`

	Status ContractStatus `json:"status"`
}

type ContractABI struct {
	Type string `json:"type" bson:"type"`
	ABI  string `json:"abi" bson:"abi"`
}

type Token struct {
	// Base information
	Address     string `json:"address" bson:"address"`
	Name        string `json:"name" bson:"name"`
	Symbol      string `json:"symbol" bson:"symbol"`
	Decimals    uint8  `json:"decimals" bson:"decimals"`
	TotalSupply string `json:"total_supply" bson:"totalSupply"`
	// Addition information
	Description       string `bson:"description"`
	Logo              string `json:"logo"`
	TotalHolders      uint   `bson:"totalHolders"`
	TotalTxs          uint   `bson:"totalTxs"`
	Price             string `bson:"price"`
	CirculatingSupply string `bson:"circulatingSupply"`
	Website           string `bson:"website"`
	Social            string `bson:"social"`
}
