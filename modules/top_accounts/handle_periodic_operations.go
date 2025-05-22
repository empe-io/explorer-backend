package top_accounts

import (
	"fmt"
	"github.com/forbole/callisto/v4/modules/utils"
	"github.com/forbole/callisto/v4/types"
	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
	"sync"
)

// RegisterPeriodicOperations implements modules.PeriodicOperationsModule
func (m *Module) RegisterPeriodicOperations(scheduler *gocron.Scheduler) error {
	log.Debug().Str("module", "top_accounts").Msg("setting up periodic tasks")

	if _, err := scheduler.Every(12).Hour().Do(func() {
		utils.WatchMethod(m.RefreshAllAccounts)
	}); err != nil {
		return fmt.Errorf("error while setting up top_accounts period operations: %s", err)
	}

	return nil
}

func (m *Module) RefreshAllAccounts() error {
	log.Debug().
		Str("module", "top_accounts").
		Str("operation", "top_accounts").
		Msg("refreshing all top accounts")

	block, err := m.db.GetLastBlockHeightAndTimestamp()
	if err != nil {
		return err
	}

	accounts, err := m.authModule.RefreshTopAccountsList(block.Height)
	if err != nil {
		return err
	}

	totalAccounts := len(accounts)
	if totalAccounts == 0 {
		log.Info().Str("module", "top_accounts").Msg("no accounts to refresh")
		return nil
	}

	// Divide accounts into 5 batches
	numBatches := 5
	batches := make([][]types.Account, 0, numBatches)
	batchSize := totalAccounts / numBatches
	remainder := totalAccounts % numBatches
	start := 0
	for i := 0; i < numBatches; i++ {
		size := batchSize
		if i < remainder {
			size++
		}
		end := start + size
		if start < totalAccounts {
			batches = append(batches, accounts[start:end])
		}
		start = end
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	totalRefreshed := 0
	errChan := make(chan error, numBatches)

	for _, batch := range batches {
		wg.Add(1)
		go func(batch []types.Account) {
			defer wg.Done()
			for _, account := range batch {
				// Process each account
				if err := m.RefreshAll(account.Address); err != nil {
					errChan <- err
					// Stop processing this batch if an error occurs
					return
				}
				mu.Lock()
				totalRefreshed++
				if totalRefreshed%100 == 0 {
					log.Info().
						Str("module", "top_accounts").
						Int("accounts_processed", totalRefreshed).
						Msg("processed 100 accounts")
				}
				mu.Unlock()
			}
		}(batch)
	}

	wg.Wait()
	close(errChan)

	// If any error occurred in the batches, return it.
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	log.Info().
		Str("module", "top_accounts").
		Int("total_accounts_refreshed", totalRefreshed).
		Msg("all accounts refreshed")
	return nil
}
