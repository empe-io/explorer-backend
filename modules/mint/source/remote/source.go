package remote

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	cfeminter "github.com/empe-io/empe-chain/x/cfeminter/types"
	mintsource "github.com/forbole/callisto/v4/modules/mint/source"
	"github.com/forbole/juno/v5/node/remote"
)

var (
	_ mintsource.Source = &Source{}
)

// Source implements mintsource.Source using a remote node
type Source struct {
	*remote.Source
	querier cfeminter.QueryClient
}

// NewSource returns a new Source instance
func NewSource(source *remote.Source, querier cfeminter.QueryClient) *Source {
	return &Source{
		Source:  source,
		querier: querier,
	}
}

// GetInflation implements mintsource.Source
func (s Source) GetInflation(height int64) (sdk.Dec, error) {
	res, err := s.querier.Inflation(remote.GetHeightRequestContext(s.Ctx, height), &cfeminter.QueryInflationRequest{})
	if err != nil {
		return sdk.ZeroDec(), err
	}
	return res.Inflation, nil
}

// Params implements mintsource.Source
func (s Source) Params(height int64) (cfeminter.Params, error) {
	res, err := s.querier.Params(remote.GetHeightRequestContext(s.Ctx, height), &cfeminter.QueryParamsRequest{})
	if err != nil {
		return cfeminter.Params{}, nil
	}

	return res.Params, nil
}
