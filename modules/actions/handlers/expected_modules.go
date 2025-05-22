package handlers

import (
	types2 "github.com/cosmos/cosmos-sdk/types"
)

type BankModule interface {
	GetLatestSupply() (types2.Coins, error)
}

type CirculatingSupplyModule interface {
	GetLatestCirculatingSupply() (int64, error)
}
