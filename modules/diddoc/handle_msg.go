package diddoc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	diddoc "github.com/empe-io/empe-chain/x/diddoc/types"
	juno "github.com/forbole/juno/v5/types"
)

// HandleMsgExec implements modules.AuthzMessageModule
func (m *Module) HandleMsgExec(index int, _ *authz.MsgExec, _ int, executedMsg sdk.Msg, tx *juno.Tx) error {
	return m.HandleMsg(index, executedMsg, tx)
}

// HandleMsg implements modules.MessageModule
func (m *Module) HandleMsg(_ int, msg sdk.Msg, tx *juno.Tx) error {
	if len(tx.Logs) == 0 {
		return nil
	}

	if casted, ok := msg.(*diddoc.MsgCreateDidDocument); ok {
		err := m.storeNewDid(casted.DidDocument, tx.Height)
		if err != nil {
			return err
		}
	}
	return nil
}
