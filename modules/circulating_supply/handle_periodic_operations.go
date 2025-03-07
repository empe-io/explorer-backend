package circulating_supply

import (
	"fmt"
	"github.com/forbole/callisto/v4/types/config"
	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"

	"github.com/forbole/callisto/v4/modules/utils"
)

// RegisterPeriodicOperations implements modules.PeriodicOperationsModule
func (m *Module) RegisterPeriodicOperations(scheduler *gocron.Scheduler) error {
	log.Debug().Str("module", "circulatingSupply").Msg("setting up periodic tasks")

	if _, err := scheduler.Every(13).Hour().Do(func() {
		utils.WatchMethod(m.UpdateCirculatingSupply)
	}); err != nil {
		return fmt.Errorf("error while setting up circulatingSupply period operations: %s", err)
	}

	return nil
}

// UpdateCirculatingSupply fetches the total amount of coins in the system from RPC and stores it in database
func (m *Module) UpdateCirculatingSupply() error {
	log.Debug().
		Str("module", "circulatingSupply").
		Str("operation", "circulatingSupply").
		Msg("updating circulating supply")

	block, err := m.db.GetLastBlockHeightAndTimestamp()
	if err != nil {
		return err
	}
	vestingAccountsLockedSum, err := m.authModule.GetCurrentlyLockedAmountSum(block.BlockTimestamp)
	if err != nil {
		return err
	}

	moduleAccoutnsSum, err := m.authModule.GetAllModuleAccountsTokensSum()
	if err != nil {
		return err
	}
	log.Debug().Str("module", "circulatingSupply").Int64("lockedTokensSum", vestingAccountsLockedSum).Int64("moduleAccountsSum", moduleAccoutnsSum).Msg("locked tokens sum")

	totalSupply, err := m.bankModule.GetSupply(block.Height)
	if err != nil {
		return err
	}
	denom := config.GetDenom()

	totalSupplyAmount := totalSupply.AmountOf(denom).Int64()
	log.Debug().Str("module", "circulatingSupply").Int64("totalSupplyAmount", totalSupplyAmount).Msg("total supply amount")

	communityPool, err := m.distrModule.GetLatestCommunityPool()
	if err != nil {
		return err
	}
	communityPoolAmount := communityPool.AmountOf(denom).TruncateInt64()
	log.Debug().Str("module", "circulatingSupply").Int64("communityPoolAmount", communityPoolAmount).Msg("community pool amount")

	circulatingSupply := totalSupplyAmount - vestingAccountsLockedSum - moduleAccoutnsSum - communityPoolAmount
	log.Debug().Str("module", "circulatingSupply").Int64("circulatingSupply", circulatingSupply).Msg("circulating supply")

	return m.db.SaveCirculatingSupply(circulatingSupply, block.Height)
}

func (m *Module) GetLatestCirculatingSupply() (int64, error) {
	return m.db.GetLatestCirculatingSupply()
}
