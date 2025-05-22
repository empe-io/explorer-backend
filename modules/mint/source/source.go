package source

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	cfeminter "github.com/empe-io/empe-chain/x/cfeminter/types"
)

type Source interface {
	GetInflation(height int64) (sdk.Dec, error)
	Params(height int64) (cfeminter.Params, error)
}
