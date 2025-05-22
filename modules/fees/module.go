package fees

import (
	"github.com/forbole/callisto/v4/database"
	"github.com/forbole/juno/v5/modules"
)

var (
	_ modules.Module = &Module{}
)

// Module represent x/top_accounts module
type Module struct {
	db *database.Db
}

// NewModule returns a new Module instance
func NewModule(
	db *database.Db,
) *Module {
	return &Module{
		db: db,
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "fees"
}
