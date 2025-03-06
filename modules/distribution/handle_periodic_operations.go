package distribution

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/types"

	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"

	"github.com/forbole/callisto/v4/modules/utils"
)

// RegisterPeriodicOperations implements modules.PeriodicOperationsModule
func (m *Module) RegisterPeriodicOperations(scheduler *gocron.Scheduler) error {
	log.Debug().Str("module", "distribution").Msg("setting up periodic tasks")

	// Update the community pool every 1 hour
	if _, err := scheduler.Every(1).Hour().Do(func() {
		utils.WatchMethod(m.UpdateLatestCommunityPool)
	}); err != nil {
		return fmt.Errorf("error while scheduling distribution periodic operation: %s", err)
	}

	return nil
}

// UpdateLatestCommunityPool gets the latest community pool from the chain and stores inside the database
func (m *Module) UpdateLatestCommunityPool() error {
	block, err := m.db.GetLastBlockHeightAndTimestamp()
	if err != nil {
		return fmt.Errorf("error while getting latest block height: %s", err)
	}

	return m.updateCommunityPool(block.Height)
}

func (m *Module) GetLatestCommunityPool() (types.DecCoins, error) {
	return m.source.GetLatestCommunityPool()
}
