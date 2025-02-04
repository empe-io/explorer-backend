package auth

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/forbole/callisto/v4/database"

	"github.com/forbole/callisto/v4/modules/auth/source"
	"github.com/forbole/juno/v5/modules"
	"github.com/forbole/juno/v5/modules/messages"
)

var (
	_ modules.Module        = &Module{}
	_ modules.GenesisModule = &Module{}
	_ modules.MessageModule = &Module{}
)

// Module represents the x/auth module
type Module struct {
	cdc            codec.Codec
	db             *database.Db
	source         source.Source
	messagesParser messages.MessageAddressesParser
}

// NewModule builds a new Module instance
func NewModule(source source.Source, messagesParser messages.MessageAddressesParser, cdc codec.Codec, db *database.Db) *Module {
	return &Module{
		messagesParser: messagesParser,
		source:         source,
		cdc:            cdc,
		db:             db,
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "auth"
}
