package diddoc

import (
	"github.com/cosmos/gogoproto/jsonpb"
	"github.com/cosmos/gogoproto/proto"
	"github.com/empe-io/empe-chain/x/diddoc/types"
	"github.com/rs/zerolog/log"
)

// updateCommunityPool fetch total amount of coins in the system from RPC and store it into database
func (m *Module) storeNewDid(didDocument *types.DidDocument, height int64) error {
	log.Debug().Str("module", "diddoc").Int64("height", height).Msg("getting community pool")
	json, err := MarshallToJson(didDocument)
	if err != nil {
		return err
	}
	// Store the pool into the database
	return m.db.SaveDidDocumentCreated(didDocument.Id, json, height)
}

func MarshallToJson(v proto.Message) ([]byte, error) {
	marshaler := &jsonpb.Marshaler{}
	jsonStr, err := marshaler.MarshalToString(v)
	return []byte(jsonStr), err
}
