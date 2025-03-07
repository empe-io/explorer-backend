package circulating_supply

import (
	types2 "github.com/cosmos/cosmos-sdk/types"
	"time"
)

type AuthModule interface {
	GetCurrentlyLockedAmountSum(currentTime time.Time) (int64, error)
	GetAllModuleAccountsTokensSum() (int64, error)
}

type BankModule interface {
	GetSupply(height int64) (types2.Coins, error)
}

type DistrModule interface {
	GetLatestCommunityPool() (types2.DecCoins, error)
}
