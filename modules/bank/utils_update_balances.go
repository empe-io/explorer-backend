package bank

import (
	"fmt"

	"github.com/forbole/callisto/v4/modules/pricefeed"
	"github.com/forbole/callisto/v4/types"
	"github.com/rs/zerolog/log"
)

// UpdateBalances updates the balances of the accounts having the given addresses,
// taking the data at the provided height
func (m *Module) UpdateBalances(addresses []string, height int64) error {
	log.Debug().Str("module", "bank").Int64("height", height).Msg("updating balances")

	balances, err := m.keeper.GetBalances(addresses, height)
	if err != nil {
		return fmt.Errorf("error while getting account balances: %s", err)
	}

	nativeTokenAmounts := make([]types.NativeTokenAmount, len(balances))
	for index, balance := range balances {
		denomAmount := balance.Balance.AmountOf(pricefeed.GetDenom())
		nativeTokenAmounts[index] = types.NewNativeTokenAmount(balance.Address, denomAmount, height)
	}

	err = m.db.SaveTopAccountsBalance("available", nativeTokenAmounts)
	if err != nil {
		return fmt.Errorf("error while saving top accounts available balances: %s", err)
	}

	return nil
}
