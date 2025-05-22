package circulating_supply

import (
	"github.com/forbole/callisto/v4/database"
	"github.com/forbole/juno/v5/modules"
)

var (
	_ modules.Module = &Module{}
)

// Module represent x/top_accounts module
type Module struct {
	db          *database.Db
	authModule  AuthModule
	bankModule  BankModule
	distrModule DistrModule
}

// NewModule returns a new Module instance
func NewModule(
	authModule AuthModule,
	bankModule BankModule,
	distrModule DistrModule,
	db *database.Db,
) *Module {
	return &Module{
		authModule:  authModule,
		bankModule:  bankModule,
		distrModule: distrModule,
		db:          db,
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "circulating_supply"
}
