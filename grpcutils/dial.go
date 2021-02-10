package grpcutils

import (
	"context"
	"time"

	"github.com/solo-io/go-utils/contextutils"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/codes"
)

// DialOpts provides a common set of options for initiating connections to a gRPC server.
type DialOpts struct {
	// address of the gRPC server
	Address string

	// connect over plaintext or HTTPS
	Insecure bool

	// Enable fast reconnection attempts on network failures (gRPC code 14 'unavailable')
	ReconnectOnNetworkFailures bool

	// Set this as the authority (host header) on the outbound dial request
	Authority string

	// additional options the caller wishes to inject
	ExtraOptions []grpc.DialOption
}

func (o DialOpts) Dial(ctx context.Context) (*grpc.ClientConn, error) {
	contextutils.LoggerFrom(ctx).Debugw("dialing grpc server", "opts", o)
	opts := []grpc.DialOption{grpc.WithBlock()}
	if o.Insecure {
		opts = append(opts, grpc.WithInsecure())
	}
	if o.ReconnectOnNetworkFailures {
		retryOpts := []grpc_retry.CallOption{
			grpc_retry.WithCodes(codes.Unavailable),
			grpc_retry.WithMax(10),
			grpc_retry.WithBackoff(grpc_retry.BackoffLinear(time.Second * 2)),
		}
		connectionBackoff := backoff.DefaultConfig
		connectionBackoff.MaxDelay = time.Second * 2
		opts = append(opts,
			grpc.WithConnectParams(grpc.ConnectParams{Backoff: connectionBackoff}),
			grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
				grpc_retry.StreamClientInterceptor(retryOpts...),
			)),
			grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
				grpc_retry.UnaryClientInterceptor(retryOpts...),
			)),
		)
	}
	if o.Authority != "" {
		opts = append(opts, grpc.WithAuthority(o.Authority))
	}
	opts = append(opts, o.ExtraOptions...)
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	cc, err := grpc.DialContext(ctx, o.Address, opts...)
	if err != nil {
		return nil, err
	}
	contextutils.LoggerFrom(ctx).Debugw("connected to grpc server", "addr", o.Address)
	return cc, nil
}
