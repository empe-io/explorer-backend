package top_accounts

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distritypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/forbole/callisto/v4/modules/utils"
	juno "github.com/forbole/juno/v5/types"
	"github.com/gogo/protobuf/proto"
)

// HandleMsg implements MessageModule
func (m *Module) HandleMsg(index int, msg sdk.Msg, tx *juno.Tx) error {
	if len(tx.Logs) == 0 {
		return nil
	}

	// Refresh x/bank available account balances
	addresses, err := m.messageParser(tx)
	if err != nil {
		return fmt.Errorf("error while parsing account addresses of message type %s: %s", proto.MessageName(msg), err)
	}

	addresses = utils.FilterNonAccountAddresses(addresses)
	err = m.bankModule.UpdateBalances(addresses, tx.Height)
	if err != nil {
		return fmt.Errorf("error while updating account available balances: %s", err)
	}

	err = m.refreshTopAccountsSum(addresses, tx.Height)
	if err != nil {
		return fmt.Errorf("error while refreshing top accounts sum while refreshing balance: %s", err)
	}

	// Handle x/staking delegations and unbondings
	switch cosmosMsg := msg.(type) {

	case *stakingtypes.MsgDelegate:
		return m.handleMsgDelegate(cosmosMsg.DelegatorAddress, tx.Height)

	case *stakingtypes.MsgUndelegate:
		return m.handleMsgUndelegate(cosmosMsg.DelegatorAddress, tx.Height)

	case *stakingtypes.MsgCancelUnbondingDelegation:
		return m.handleMsgCancelUnbondingDelegation(cosmosMsg.DelegatorAddress, tx.Height)

	// Handle x/distribution delegator rewards
	case *distritypes.MsgWithdrawDelegatorReward:
		return m.handleMsgWithdrawDelegatorReward(cosmosMsg.DelegatorAddress, tx.Height)

	}

	return nil
}

func (m *Module) handleMsgDelegate(delAddr string, height int64) error {
	err := m.stakingModule.RefreshDelegations(delAddr, height)
	if err != nil {
		return fmt.Errorf("error while refreshing delegations while handling MsgDelegate: %s", err)
	}

	err = m.refreshTopAccountsSum([]string{delAddr}, height)
	if err != nil {
		return fmt.Errorf("error while refreshing top accounts sum while handling MsgDelegate: %s", err)
	}

	return nil
}

// handleMsgUndelegate handles a MsgUndelegate storing the data inside the database
func (m *Module) handleMsgUndelegate(delAddr string, height int64) error {
	err := m.stakingModule.RefreshUnbondings(delAddr, height)
	if err != nil {
		return fmt.Errorf("error while refreshing undelegations while handling MsgUndelegate: %s", err)
	}

	err = m.refreshTopAccountsSum([]string{delAddr}, height)
	if err != nil {
		return fmt.Errorf("error while refreshing top accounts sum while handling MsgUndelegate: %s", err)
	}

	return nil
}

// handleMsgCancelUnbondingDelegation handles a MsgCancelUnbondingDelegation storing the data inside the database
func (m *Module) handleMsgCancelUnbondingDelegation(delAddr string, height int64) error {
	err := m.stakingModule.RefreshDelegations(delAddr, height)
	if err != nil {
		return fmt.Errorf("error while refreshing delegations of account %s, error: %s", delAddr, err)
	}

	err = m.stakingModule.RefreshUnbondings(delAddr, height)
	if err != nil {
		return fmt.Errorf("error while refreshing unbonding delegations of account %s, error: %s", delAddr, err)
	}

	err = m.bankModule.UpdateBalances([]string{delAddr}, height)
	if err != nil {
		return fmt.Errorf("error while refreshing balance of account %s, error: %s", delAddr, err)
	}

	err = m.refreshTopAccountsSum([]string{delAddr}, height)
	if err != nil {
		return fmt.Errorf("error while refreshing top accounts sum %s, error: %s", delAddr, err)
	}

	return nil
}

func (m *Module) handleMsgWithdrawDelegatorReward(delAddr string, height int64) error {
	err := m.distrModule.RefreshDelegatorRewards([]string{delAddr}, height)
	if err != nil {
		return fmt.Errorf("error while refreshing delegator rewards: %s", err)
	}

	err = m.bankModule.UpdateBalances([]string{delAddr}, height)
	if err != nil {
		return fmt.Errorf("error while updating account available balances with MsgWithdrawDelegatorReward: %s", err)
	}

	err = m.refreshTopAccountsSum([]string{delAddr}, height)
	if err != nil {
		return fmt.Errorf("error while refreshing top accounts sum while handling MsgWithdrawDelegatorReward: %s", err)
	}

	return nil
}
