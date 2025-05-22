package remote

import (
	"context"
	"crypto/tls"
	"fmt"
	params2 "github.com/empe-io/empe-chain/app/params"
	"github.com/forbole/juno/v5/node/remote"
	"net"
	"regexp"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	proto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

var (
	HTTPProtocols = regexp.MustCompile("https?://")
)

// GRPCConfig holds the configuration for a gRPC connection.
type GRPCConfig struct {
	Address  string
	Insecure bool
}

// GetHeightRequestContext adds the height to the context for querying the state at a given height.
func GetHeightRequestContext(ctx context.Context, height int64) context.Context {
	return metadata.AppendToOutgoingContext(
		ctx,
		"x-cosmos-block-height",
		strconv.FormatInt(height, 10),
	)
}

// customCodec wraps a parent codec to provide custom marshaling and unmarshaling.
type customCodec struct {
	parentCodec codec.Codec
}

// Marshal converts a proto.Message into bytes using gogoproto.
func (c customCodec) Marshal(v interface{}) ([]byte, error) {
	protoMsg, ok := v.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("failed to assert proto.Message")
	}
	return proto.Marshal(protoMsg)
}

// Unmarshal converts bytes into a proto.Message using gogoproto.
func (c customCodec) Unmarshal(data []byte, v interface{}) error {
	protoMsg, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("failed to assert proto.Message")
	}
	return proto.Unmarshal(data, protoMsg)
}

// Name returns the name of the custom codec.
func (c customCodec) Name() string {
	return "gogoproto"
}

// MustCreateGrpcConnection creates a new gRPC connection using the provided configuration
// and panics on error.
func MustCreateGrpcConnection(cfg *remote.GRPCConfig) *grpc.ClientConn {
	conn, err := CreateGrpcConnection(cfg)
	if err != nil {
		panic(err)
	}
	return conn
}

// CreateGrpcConnection creates a new gRPC client connection from the given configuration
// with custom codec support.
func CreateGrpcConnection(cfg *remote.GRPCConfig) (*grpc.ClientConn, error) {
	var grpcOpts []grpc.DialOption

	// Initialize custom codec (adjust parentCodec as needed)
	customCodecInstance := customCodec{
		parentCodec: params2.MakeEncodingConfig().Marshaler,
	}

	// Add custom codec option to force usage of our codec.
	callOptions := grpc.WithDefaultCallOptions(grpc.ForceCodec(customCodecInstance))

	// Set up security options.
	if cfg.Insecure {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}

	// Append the custom codec call option.
	grpcOpts = append(grpcOpts, callOptions)

	// Remove any http(s) prefix from the address.
	address := HTTPProtocols.ReplaceAllString(cfg.Address, "")
	return grpc.Dial(address, grpcOpts...)
}

// Client is an example client struct holding a gRPC connection.
type Client struct {
	tls            bool
	grpcEndpoint   string
	grpcConnection *grpc.ClientConn
}

// dial is a custom dialer for gRPC connections (adjust as needed).
func dial(ctx context.Context, addr string) (net.Conn, error) {
	// In this example we use TLS dialing. Modify as necessary.
	return tls.Dial("tcp", addr, &tls.Config{})
}

// ConnectGRPC dials the gRPC connection endpoint using the custom codec.
func (c *Client) ConnectGRPC() error {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}

	customCodecInstance := customCodec{
		parentCodec: params2.MakeEncodingConfig().Marshaler,
	}

	var grpcOpts []grpc.DialOption
	if c.tls {
		grpcOpts = []grpc.DialOption{
			grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
			grpc.WithContextDialer(dial),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{}),
			grpc.WithDefaultCallOptions(grpc.ForceCodec(customCodecInstance)),
		}
	} else {
		grpcOpts = []grpc.DialOption{
			grpc.WithContextDialer(dial),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(grpc.ForceCodec(customCodecInstance)),
		}
	}

	grpcConn, err := grpc.Dial(c.grpcEndpoint, grpcOpts...)
	if err != nil {
		return fmt.Errorf("failed to dial Cosmos gRPC service: %w", err)
	}

	c.grpcConnection = grpcConn
	return nil
}
