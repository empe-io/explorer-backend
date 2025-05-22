package mint

import (
	"encoding/json"
	"fmt"
	"github.com/CosmWasm/wasmd/app"

	tmtypes "github.com/cometbft/cometbft/types"

	cfemintertypes "github.com/empe-io/empe-chain/x/cfeminter/types"
	"github.com/forbole/callisto/v4/types"

	"github.com/rs/zerolog/log"
)

// HandleGenesis implements modules.Module
func (m *Module) HandleGenesis(doc *tmtypes.GenesisDoc, appState map[string]json.RawMessage) error {
	log.Debug().Str("module", "mint").Msg("parsing genesis")

	// Read the genesis state
	var genState cfemintertypes.GenesisState
	err := app.MakeEncodingConfig().Codec.UnmarshalJSON(appState[cfemintertypes.ModuleName], &genState)
	if err != nil {
		return fmt.Errorf("error while reading mint genesis data: %s", err)
	}

	// Save the params
	err = m.db.SaveMintParams(types.NewMintParams(genState.Params, doc.InitialHeight))
	if err != nil {
		return fmt.Errorf("error while storing genesis mint params: %s", err)
	}

	return nil
}
