package types

type Address struct {
	Address       string  `json:"address" bson:"address"`
	Rank          uint64  `json:"rank"`
	BalanceFloat  float64 `json:"-" bson:"balanceFloat"`        // low precise balance for sorting purposes
	BalanceString string  `json:"balance" bson:"balanceString"` // high precise balance for API
	Name          string  `json:"name" bson:"name"`             // alias of an address

	// Token
	TokenName   string `json:"tokenName" bson:"tokenName"`
	TokenSymbol string `json:"tokenSymbol" bson:"tokenSymbol"`
	Decimals    int64  `json:"decimals" bson:"decimals"`
	TotalSupply string `json:"totalSupply" bson:"totalSupply"`

	// SMC
	IsContract   bool     `json:"isContract" bson:"isContract"`
	ErcTypes     []string `json:"ercTypes" bson:"ercTypes"`
	Interfaces   []string `json:"interfaces" bson:"interfaces"`
	OwnerAddress string   `json:"ownerAddress" bson:"ownerAddress"`

	// Stats
	TxCount         int `json:"txCount" bson:"txCount"`
	HolderCount     int `json:"holderCount" bson:"holderCount"`
	InternalTxCount int `json:"internalTxCount" bson:"internalTxCount"`
	TokenTxCount    int `json:"tokenTxCount" bson:"tokenTxCount"`

	UpdatedAt int64 `json:"updatedAt" bson:"updatedAt"`
}

type UpdateAddress struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}
