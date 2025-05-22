package database

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"

	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/lib/pq"

	dbtypes "github.com/forbole/callisto/v4/database/types"
	dbutils "github.com/forbole/callisto/v4/database/utils"

	"github.com/forbole/callisto/v4/types"
)

// SaveAccounts saves the given accounts inside the database
func (db *Db) SaveAccounts(accounts []types.Account) error {
	paramsNumber := 1
	slices := dbutils.SplitAccounts(accounts, paramsNumber)

	for _, accounts := range slices {
		if len(accounts) == 0 {
			continue
		}

		// Store up-to-date data
		err := db.saveAccounts(paramsNumber, accounts)
		if err != nil {
			return fmt.Errorf("error while storing accounts: %s", err)
		}
	}

	return nil
}

func (db *Db) saveAccounts(paramsNumber int, accounts []types.Account) error {
	if len(accounts) == 0 {
		return nil
	}

	stmt := `INSERT INTO account (address) VALUES `
	var params []interface{}

	for i, account := range accounts {
		ai := i * paramsNumber
		stmt += fmt.Sprintf("($%d),", ai+1)
		params = append(params, account.Address)
	}

	stmt = stmt[:len(stmt)-1]
	stmt += " ON CONFLICT DO NOTHING"
	_, err := db.SQL.Exec(stmt, params...)
	if err != nil {
		return fmt.Errorf("error while storing accounts: %s", err)
	}

	return nil
}

func (db *Db) SaveVestingAccount(account exported.VestingAccount) error {
	switch vestingAccount := account.(type) {
	case *vestingtypes.ContinuousVestingAccount, *vestingtypes.DelayedVestingAccount:
		_, err := db.storeVestingAccount(account)
		if err != nil {
			return err
		}

	case *vestingtypes.PeriodicVestingAccount:
		vestingAccountRowID, err := db.storeVestingAccount(account)
		if err != nil {
			return err
		}
		err = db.storeVestingPeriods(vestingAccountRowID, vestingAccount.VestingPeriods)
		if err != nil {
			return err
		}
	}

	return nil
}

// SaveVestingAccounts saves the given vesting accounts inside the database
func (db *Db) SaveVestingAccounts(vestingAccounts []exported.VestingAccount) error {
	if len(vestingAccounts) == 0 {
		return nil
	}

	for _, account := range vestingAccounts {
		err := db.SaveVestingAccount(account)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Db) GetCurrentlyLockedAmountSum(currentTime time.Time) (int64, error) {
	var total int64
	query := `
		SELECT COALESCE(SUM(locked), 0)::bigint AS total_locked
		FROM (
			SELECT CASE
				WHEN $1 <= start_time THEN coin.amount::numeric
				WHEN $1 >= end_time THEN 0
				ELSE coin.amount::numeric - ROUND(coin.amount::numeric * ((EXTRACT(EPOCH FROM $1) - EXTRACT(EPOCH FROM start_time)) / (EXTRACT(EPOCH FROM end_time) - EXTRACT(EPOCH FROM start_time))))
			END AS locked
			FROM vesting_account,
			LATERAL unnest(original_vesting) AS coin
			WHERE coin.denom = 'uempe'
		) AS sub;
	`
	err := db.Sqlx.Get(&total, query, currentTime)
	return total, err
}

func (db *Db) GetAllModuleAccountsTokensSum() (int64, error) {
	var sum int64
	err := db.Sqlx.Get(&sum, `
		SELECT COALESCE(SUM("sum"), 0)
		FROM top_accounts
		WHERE type = 'cosmos.vesting.v1beta1.ModuleAccount'
	`)
	return sum, err
}

func (db *Db) GetAvailableTokensSum(addresses []string) (int64, error) {
	var sum int64
	query, args, err := sqlx.In(`
		SELECT COALESCE(SUM(available), 0)
		FROM top_accounts
		WHERE address IN (?)
	`, addresses)
	if err != nil {
		return 0, err
	}
	query = db.Sqlx.Rebind(query)
	err = db.Sqlx.Get(&sum, query, args...)
	return sum, err
}

func (db *Db) GetAllVestingAccounts() ([]exported.VestingAccount, error) {
	var rows []exported.VestingAccount
	err := db.Sqlx.Select(&rows, `SELECT * FROM vesting_account`)
	return rows, err
}

func (db *Db) storeVestingAccount(account exported.VestingAccount) (int, error) {
	stmt := `
	INSERT INTO vesting_account (type, address, original_vesting, end_time, start_time) 
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (address) DO UPDATE 
		SET original_vesting = excluded.original_vesting, 
			end_time = excluded.end_time, 
			start_time = excluded.start_time
			RETURNING id `

	// Store the vesting account
	err := db.SaveAccounts([]types.Account{types.NewAccount(account.GetAddress().String())})
	if err != nil {
		return 0, fmt.Errorf("error while storing vesting account address: %s", err)
	}

	var vestingAccountRowID int
	err = db.SQL.QueryRow(stmt,
		proto.MessageName(account),
		account.GetAddress().String(),
		pq.Array(dbtypes.NewDbCoins(account.GetOriginalVesting())),
		time.Unix(account.GetEndTime(), 0),
		time.Unix(account.GetStartTime(), 0),
	).Scan(&vestingAccountRowID)

	if err != nil {
		return vestingAccountRowID, fmt.Errorf("error while saving Vesting Account of type %v: %s", proto.MessageName(account), err)
	}

	return vestingAccountRowID, nil
}

func (db *Db) StoreBaseVestingAccountFromMsg(bva *vestingtypes.BaseVestingAccount, txTimestamp time.Time) error {
	stmt := `
	INSERT INTO vesting_account (type, address, original_vesting, start_time, end_time) 
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (address) DO UPDATE 
		SET type = excluded.type,
			original_vesting = excluded.original_vesting, 
			start_time = excluded.start_time, 
			end_time = excluded.end_time`

	// Store the vesting account
	err := db.SaveAccounts([]types.Account{types.NewAccount(bva.GetAddress().String())})
	if err != nil {
		return fmt.Errorf("error while storing vesting account address: %s", err)
	}

	_, err = db.SQL.Exec(stmt,
		proto.MessageName(bva),
		bva.GetAddress().String(),
		pq.Array(dbtypes.NewDbCoins(bva.OriginalVesting)),
		txTimestamp,
		time.Unix(bva.EndTime, 0))
	if err != nil {
		return fmt.Errorf("error while storing vesting account: %s", err)
	}
	return nil
}

// storeVestingPeriods handles storing the vesting periods of PeriodicVestingAccount type
func (db *Db) storeVestingPeriods(id int, vestingPeriods []vestingtypes.Period) error {
	// Delete already existing periods
	stmt := `DELETE FROM vesting_period WHERE vesting_account_id = $1`
	_, err := db.SQL.Exec(stmt, id)
	if err != nil {
		return fmt.Errorf("error while deleting vesting period: %s", err)
	}

	// Store the new periods
	stmt = `
INSERT INTO vesting_period (vesting_account_id, period_order, length, amount) 
VALUES `

	var params []interface{}
	for i, period := range vestingPeriods {
		ai := i * 4
		stmt += fmt.Sprintf("($%d,$%d,$%d,$%d),", ai+1, ai+2, ai+3, ai+4)

		order := i
		amount := pq.Array(dbtypes.NewDbCoins(period.Amount))
		params = append(params, id, order, period.Length, amount)
	}
	stmt = stmt[:len(stmt)-1]

	_, err = db.SQL.Exec(stmt, params...)
	if err != nil {
		return fmt.Errorf("error while saving vesting periods: %s", err)
	}

	return nil
}

// GetAccounts returns all the accounts that are currently stored inside the database.
func (db *Db) GetAccounts() ([]string, error) {
	var rows []string
	err := db.Sqlx.Select(&rows, `SELECT address FROM account`)
	return rows, err
}
