package database

import (
	"fmt"
)

// SaveFees stores a new fees record at the given height,
// inserting feeValue and stableFeeValue. If a record for that height already exists,
// it updates the record with the new values.
func (db *Db) SaveFees(feeValue int64, stableFeeValue int64, height int64) error {
	stmt := `
INSERT INTO fees (height, fee_value, stable_fee_value)
VALUES ($1, $2, $3)
ON CONFLICT (height) DO UPDATE 
    SET fee_value = EXCLUDED.fee_value,
        stable_fee_value = EXCLUDED.stable_fee_value;
`
	_, err := db.SQL.Exec(stmt, height, feeValue, stableFeeValue)
	if err != nil {
		return fmt.Errorf("error while storing fees: %s", err)
	}

	return nil
}

// GetLatestFeesHeight retrieves the latest fees record by height.
// It returns both the fee value and the height at which it was recorded.
func (db *Db) GetLatestFeesHeight() (int64, error) {
	stmt := `SELECT COALESCE(MAX(height), 0) AS height FROM fees`
	var height int64
	err := db.SQL.Get(&height, stmt)
	if err != nil {
		return 0, fmt.Errorf("error fetching latest fees height: %w", err)
	}
	return height, nil
}

// GetTransactionsAggregates retrieves all transactions between fromHeight and toHeight,
// and aggregates the following from those transactions:
//  1. The sum of all fee amounts (from the fee JSONB column) where the fee "amount" array contains objects with "denom" equal to 'uempe'.
//  2. The sum of all event values (from the logs JSONB column) where the event type is 'empe.stablefee.EventChargeFee'
//     and the event has an attribute with key 'uempeAmount'.
//
// Note: Each transaction may contain multiple fee entries or events, and the sums are computed accordingly.
func (db *Db) GetTransactionsAggregates(fromHeight, toHeight int64) (int64, int64, error) {
	aggQuery := `
WITH tx AS (
    SELECT fee, logs
    FROM transaction
    WHERE height BETWEEN $1 AND $2
),
fee_total AS (
    SELECT COALESCE(SUM((replace(fee_elem->>'amount', '"', '')::numeric)), 0) AS total_fee
    FROM tx,
         jsonb_array_elements(fee->'amount') AS fee_elem
    WHERE fee_elem->>'denom' = 'uempe'
),
event_total AS (
    SELECT COALESCE(SUM((replace(attr->>'value', '"', '')::numeric)), 0) AS total_event
    FROM tx,
         jsonb_array_elements(logs) AS log,
         jsonb_array_elements(log->'events') AS event,
         jsonb_array_elements(event->'attributes') AS attr
    WHERE event->>'type' = 'empe.stablefee.EventChargeFee'
      AND attr->>'key' = 'uempeAmount'
)
SELECT 
    (SELECT total_fee FROM fee_total) AS total_fees,
    (SELECT total_event FROM event_total) AS total_event_charge;
`
	var agg struct {
		TotalFees        float64 `db:"total_fees"`
		TotalEventCharge float64 `db:"total_event_charge"`
	}
	err := db.SQL.Get(&agg, aggQuery, fromHeight, toHeight)
	if err != nil {
		return 0, 0, fmt.Errorf("error fetching aggregates: %s", err)
	}

	return int64(agg.TotalFees), int64(agg.TotalEventCharge), nil
}
