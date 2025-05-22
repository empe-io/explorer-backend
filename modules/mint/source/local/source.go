package local

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cfeminter "github.com/empe-io/empe-chain/x/cfeminter/types"
	mintsource "github.com/forbole/callisto/v4/modules/mint/source"
	"github.com/forbole/juno/v5/node/local"
)

var (
	_ mintsource.Source = &Source{}
)

// Source implements mintsource.Source using a local node
type Source struct {
	*local.Source
	querier cfeminter.QueryServer
}

// NewSource returns a new Source instace
func NewSource(source *local.Source, querier cfeminter.QueryServer) *Source {
	return &Source{
		Source:  source,
		querier: querier,
	}
}

// GetInflation implements mintsource.Source
func (s Source) GetInflation(height int64) (sdk.Dec, error) {
	ctx, err := s.LoadHeight(height)
	if err != nil {
		return sdk.ZeroDec(), fmt.Errorf("error while loading height: %s", err)
	}
	res, err := s.querier.Inflation(sdk.WrapSDKContext(ctx), &cfeminter.QueryInflationRequest{})
	if err != nil {
		return sdk.ZeroDec(), err
	}
	return res.Inflation, nil
}

// Params implements mintsource.Source
func (s Source) Params(height int64) (cfeminter.Params, error) {
	ctx, err := s.LoadHeight(height)
	if err != nil {
		return cfeminter.Params{}, fmt.Errorf("error while loading height: %s", err)
	}

	res, err := s.querier.Params(sdk.WrapSDKContext(ctx), &cfeminter.QueryParamsRequest{})
	if err != nil {
		return cfeminter.Params{}, err
	}

	return res.Params, nil
}
