package types

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	empe "github.com/empe-io/empe-chain/app"
	params2 "github.com/empe-io/empe-chain/app/params"
	cfemintertypes "github.com/empe-io/empe-chain/x/cfeminter/types"
	"os"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/forbole/juno/v5/node/remote"
	"github.com/forbole/juno/v5/types/params"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/forbole/juno/v5/node/local"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	authsource "github.com/forbole/callisto/v4/modules/auth/source"
	localauthsource "github.com/forbole/callisto/v4/modules/auth/source/local"
	remoteauthsource "github.com/forbole/callisto/v4/modules/auth/source/remote"
	banksource "github.com/forbole/callisto/v4/modules/bank/source"
	localbanksource "github.com/forbole/callisto/v4/modules/bank/source/local"
	remotebanksource "github.com/forbole/callisto/v4/modules/bank/source/remote"
	distrsource "github.com/forbole/callisto/v4/modules/distribution/source"
	remotedistrsource "github.com/forbole/callisto/v4/modules/distribution/source/remote"
	govsource "github.com/forbole/callisto/v4/modules/gov/source"
	localgovsource "github.com/forbole/callisto/v4/modules/gov/source/local"
	remotegovsource "github.com/forbole/callisto/v4/modules/gov/source/remote"

	mintsource "github.com/forbole/callisto/v4/modules/mint/source"
	localmintsource "github.com/forbole/callisto/v4/modules/mint/source/local"
	remotemintsource "github.com/forbole/callisto/v4/modules/mint/source/remote"
	slashingsource "github.com/forbole/callisto/v4/modules/slashing/source"
	localslashingsource "github.com/forbole/callisto/v4/modules/slashing/source/local"
	remoteslashingsource "github.com/forbole/callisto/v4/modules/slashing/source/remote"
	stakingsource "github.com/forbole/callisto/v4/modules/staking/source"
	localstakingsource "github.com/forbole/callisto/v4/modules/staking/source/local"
	remotestakingsource "github.com/forbole/callisto/v4/modules/staking/source/remote"
	nodeconfig "github.com/forbole/juno/v5/node/config"
)

type Sources struct {
	AuthSource     authsource.Source
	BankSource     banksource.Source
	DistrSource    distrsource.Source
	GovSource      govsource.Source
	MintSource     mintsource.Source
	SlashingSource slashingsource.Source
	StakingSource  stakingsource.Source
}

func BuildSources(nodeCfg nodeconfig.Config, encodingConfig params.EncodingConfig) (*Sources, error) {
	switch cfg := nodeCfg.Details.(type) {
	case *remote.Details:
		return buildRemoteSources(cfg)
	case *local.Details:
		return buildLocalSources(cfg, encodingConfig)

	default:
		return nil, fmt.Errorf("invalid configuration type: %T", cfg)
	}
}

func buildLocalSources(cfg *local.Details, encodingConfig params.EncodingConfig) (*Sources, error) {
	source, err := local.NewSource(cfg.Home, encodingConfig)
	if err != nil {
		return nil, err
	}

	app := empe.NewEmpeApp(
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
		source.StoreDB, nil, true, nil, cfg.Home, params2.MakeEncodingConfig(), sims.EmptyAppOptions{}, []wasmkeeper.Option{},
	)

	sources := &Sources{
		BankSource: localbanksource.NewSource(source, banktypes.QueryServer(app.BankKeeper)),
		AuthSource: localauthsource.NewSource(source, authtypes.QueryServer(app.AccountKeeper)),
		// DistrSource:    localdistrsource.NewSource(source, distrtypes.QueryServer(app.DistrKeeper)),
		GovSource:      localgovsource.NewSource(source, govtypesv1.QueryServer(app.GovKeeper)),
		MintSource:     localmintsource.NewSource(source, cfemintertypes.QueryServer(app.CfeminterKeeper)),
		SlashingSource: localslashingsource.NewSource(source, slashingtypes.QueryServer(app.SlashingKeeper)),
		StakingSource:  localstakingsource.NewSource(source, stakingkeeper.Querier{Keeper: app.StakingKeeper}),
	}

	// Mount and initialize the stores
	err = source.MountKVStores(app, "keys")
	if err != nil {
		return nil, err
	}

	err = source.MountTransientStores(app, "tkeys")
	if err != nil {
		return nil, err
	}

	err = source.MountMemoryStores(app, "memKeys")
	if err != nil {
		return nil, err
	}

	err = source.InitStores()
	if err != nil {
		return nil, err
	}

	return sources, nil
}

func buildRemoteSources(cfg *remote.Details) (*Sources, error) {
	source, err := remote.NewSource(cfg.GRPC)
	source.GrpcConn = remotemintsource.MustCreateGrpcConnection(cfg.GRPC)
	if err != nil {
		return nil, fmt.Errorf("error while creating remote source: %s", err)
	}

	return &Sources{
		AuthSource:     remoteauthsource.NewSource(source, authtypes.NewQueryClient(source.GrpcConn)),
		BankSource:     remotebanksource.NewSource(source, banktypes.NewQueryClient(source.GrpcConn)),
		DistrSource:    remotedistrsource.NewSource(source, distrtypes.NewQueryClient(source.GrpcConn)),
		GovSource:      remotegovsource.NewSource(source, govtypesv1.NewQueryClient(source.GrpcConn)),
		MintSource:     remotemintsource.NewSource(source, cfemintertypes.NewQueryClient(source.GrpcConn)),
		SlashingSource: remoteslashingsource.NewSource(source, slashingtypes.NewQueryClient(source.GrpcConn)),
		StakingSource:  remotestakingsource.NewSource(source, stakingtypes.NewQueryClient(source.GrpcConn)),
	}, nil
}
