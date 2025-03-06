package database

import (
	"fmt"
)

// SaveCirculatingSupply allows to store the circulating supply at the given height
func (db *Db) SaveCirculatingSupply(circulatingSupply int64, height int64) error {
	stmt := `
INSERT INTO circulating_supply (value, height) 
VALUES ($1, $2) 
ON CONFLICT (one_row_id) DO UPDATE 
    SET value = excluded.value, 
        height = excluded.height 
WHERE circulating_supply.height <= excluded.height`

	_, err := db.SQL.Exec(stmt, circulatingSupply, height)
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
