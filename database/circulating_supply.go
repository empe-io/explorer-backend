package database

import (
	"fmt"
)

// SaveCirculatingSupply stores a circulating supply record at the given height.
// If a record for that height already exists, it updates the supply value.
func (db *Db) SaveCirculatingSupply(circulatingSupply int64, height int64) error {
	stmt := `
INSERT INTO circulating_supply (height, value)
VALUES ($1, $2)
ON CONFLICT (height) DO UPDATE 
    SET value = EXCLUDED.value;
`
	_, err := db.SQL.Exec(stmt, height, circulatingSupply)
	if err != nil {
		return fmt.Errorf("error while storing circulatingSupply: %s", err)
	}

	return nil
}

func (db *Db) GetLatestCirculatingSupply() (int64, error) {
	stmt := `SELECT value FROM circulating_supply ORDER BY height DESC LIMIT 1`
	var circulatingSupply int64
	err := db.SQL.Get(&circulatingSupply, stmt)
	if err != nil {
		return 0, fmt.Errorf("error while getting circulatingSupply: %s", err)
	}

	return circulatingSupply, nil
}
