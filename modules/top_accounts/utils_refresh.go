package top_accounts

import (
	"fmt"
)

func (m *Module) refreshTopAccountsSum(addresses []string, height int64) error {
	for _, addr := range addresses {
		sum, err := m.db.GetAccountBalanceSum(addr)
		if err != nil {
			return fmt.Errorf("error while getting account balance sum : %s", err)
		}

		err = m.db.UpdateTopAccountsSum(addr, sum, height)
		if err != nil {
			return fmt.Errorf("error while updating top accounts sum : %s", err)
		}
	}
	return nil
}
