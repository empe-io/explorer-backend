package database

import (
	"fmt"
)

// SaveCommunityPool allows to save for the given height the given total amount of coins
func (db *Db) SaveDidDocumentCreated(did string, json []byte, height int64) error {
	query := `
INSERT INTO did_document(did, height, json) 
VALUES ($1, $2, $3)`
	_, err := db.SQL.Exec(query, did, height, json)
	if err != nil {
		return fmt.Errorf("error while storing did document: %s", err)
	}

	return nil
}
