package circulating_supply

import (
	"fmt"
	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"

	"github.com/forbole/callisto/v4/modules/utils"
)

const DENOM = "uempe"

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
		Str("module", "circulating_supply").
		Str("operation", "circulating_supply").
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
	log.Debug().Str("module", "circulating_supply").Int64("lockedTokensSum", vestingAccountsLockedSum).Int64("moduleAccountsSum", moduleAccoutnsSum)

	totalSupply, err := m.bankModule.GetSupply(block.Height)
	if err != nil {
		return err
	}
	totalSupplyAmount := totalSupply.AmountOf(DENOM).Int64()
	log.Debug().Str("module", "circulating_supply").Int64("totalSupplyAmount", totalSupplyAmount)

	communityPool, err := m.distrModule.GetLatestCommunityPool()
	if err != nil {
		return err
	}
	communityPoolAmount := communityPool.AmountOf(DENOM).TruncateInt64()
	log.Debug().Str("module", "circulating_supply").Int64("communityPoolAmount", communityPoolAmount)

	circulatingSupply := totalSupplyAmount - vestingAccountsLockedSum - moduleAccoutnsSum - communityPoolAmount
	log.Debug().Str("module", "circulating_supply").Int64("circulatingSupply", circulatingSupply)

	return m.db.SaveCirculatingSupply(circulatingSupply, block.Height)
}

func (m *Module) GetLatestCirculatingSupply() (int64, error) {
	return m.db.GetLatestCirculatingSupply()
}
