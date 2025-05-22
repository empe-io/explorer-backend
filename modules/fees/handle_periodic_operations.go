package fees

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"

	"github.com/forbole/callisto/v4/modules/utils"
)

// RegisterPeriodicOperations implements modules.PeriodicOperationsModule
func (m *Module) RegisterPeriodicOperations(scheduler *gocron.Scheduler) error {
	log.Debug().Str("module", "fees").Msg("setting up periodic tasks")

	if _, err := scheduler.Every(6).Hours().Do(func() {
		utils.WatchMethod(m.UpdateCollectedFees)
	}); err != nil {
		return fmt.Errorf("error while setting up fees period operations: %s", err)
	}

	return nil
}

// UpdateCollectedFees updates the collected fees
func (m *Module) UpdateCollectedFees() error {
	log.Debug().
		Str("module", "fees").
		Str("operation", "update_collected_fees").
		Msg("updating collected fees")

	block, err := m.db.GetLastBlockHeightAndTimestamp()
	if err != nil {
		return err
	}

	latestSavedFeesHeight, err := m.GetLatestFeesHeight()
	log.Debug().Str("module", "fees").Bool("errors.Is(err, sql.ErrNoRows)", errors.Is(err, sql.ErrNoRows))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if latestSavedFeesHeight == block.Height {
		log.Debug().Str("module", "fees").Int64("latestSavedFeesHeight", latestSavedFeesHeight).Msg("fees already updated")
		return nil
	}

	log.Debug().Str("module", "fees").Int64("latestSavedFeesHeight", latestSavedFeesHeight).Int64("block.Height", block.Height).Msg("updating fees")
	totalFees, totalStablefeeFees, err := m.db.GetTransactionsAggregates(latestSavedFeesHeight+1, block.Height)
	if err != nil {
		return err
	}
	log.Debug().Str("module", "fees").Int64("totalFees", totalFees).Int64("totalStablefeeFees", totalStablefeeFees).Msg("total fees")

	return m.db.SaveFees(totalFees, totalStablefeeFees, block.Height)
}

func (m *Module) GetLatestFeesHeight() (height int64, error error) {
	return m.db.GetLatestFeesHeight()
}
