package top_accounts

import (
	"fmt"
	eventsutil "github.com/forbole/callisto/v4/utils/events"

	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/forbole/juno/v5/types"
)

// HandleMsg implements BlockModule
func (m *Module) HandleBlock(block *tmctypes.ResultBlock, results *tmctypes.ResultBlockResults, txs []*types.Tx, vals *tmctypes.ResultValidators) error {
	// handle complete unbonding event
	events := sdk.StringifyEvents(results.EndBlockEvents)
	height := block.Block.Height

	completeUnbondingEvents := eventsutil.FilterEvents(events, stakingtypes.EventTypeCompleteUnbonding)

	for _, event := range completeUnbondingEvents {
		// get the delegator address
		delAttr, found := eventsutil.FindAttributeByKey(event, stakingtypes.AttributeKeyDelegator)
		if !found {
			continue
		}

		address := delAttr.Value
		err := m.stakingModule.RefreshDelegations(address, height)
		if err != nil {
			return fmt.Errorf("error while refreshing delegations of account %s, error: %s", address, err)
		}

		err = m.stakingModule.RefreshUnbondings(address, height)
		if err != nil {
			return fmt.Errorf("error while refreshing unbonding delegations of account %s, error: %s", address, err)
		}

		err = m.bankModule.UpdateBalances([]string{address}, height)
		if err != nil {
			return fmt.Errorf("error while refreshing balance of account %s, error: %s", address, err)
		}

		err = m.refreshTopAccountsSum([]string{address}, height)
		if err != nil {
			return fmt.Errorf("error while refreshing top accounts sum %s, error: %s", address, err)
		}
	}

	return nil
}
