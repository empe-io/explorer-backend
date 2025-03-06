package circulating_supply

import (
	types2 "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/callisto/v4/types"
	"time"
)

type AuthModule interface {
	GetAllBaseAccounts(height int64) ([]types.Account, error)
	RefreshTopAccountsList(height int64) ([]types.Account, error)
	GetCurrentlyLockedAmountSum(currentTime time.Time) (int64, error)
	GetAvailableTokensSum(addresses []string) (int64, error)
	GetAllModuleAccountsTokensSum() (int64, error)
}

type AuthSource interface {
	GetTotalNumberOfAccounts(height int64) (uint64, error)
}

type BankModule interface {
	UpdateBalances(addresses []string, height int64) error
	GetSupply(height int64) (types2.Coins, error)
}

type DistrModule interface {
	RefreshDelegatorRewards(delegators []string, height int64) error
	GetLatestCommunityPool() (types2.DecCoins, error)
}

type StakingModule interface {
	RefreshDelegations(delegatorAddr string, height int64) error
	RefreshUnbondings(delegatorAddr string, height int64) error
}
