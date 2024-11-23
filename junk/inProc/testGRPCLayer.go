package inProc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/madflojo/testcerts"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
	"strings"
	"sync"
	"testing"
	"time"
)

type TestGRPCLayer struct {
	t                 *testing.T
	physicalTransport *Network
	publicKeyInfra    *testcerts.CertificateAuthority

	//todo: resolver is a mess in terms of initial seeds versus actual.  might be better to build out a custom binding
	//to avoid needing to update.
	resolverControl sync.Mutex
	resolverState   resolver.State
	resolver        *manual.Resolver
}

func NewTestGRPCLayer(t *testing.T) *TestGRPCLayer {
	return &TestGRPCLayer{
		t:                 t,
		publicKeyInfra:    testcerts.NewCA(),
		physicalTransport: NewNetwork(),
		resolverControl:   sync.Mutex{},
		resolverState:     resolver.State{},
		resolver:          manual.NewBuilderWithScheme("in-proc"),
	}
}

func (t *TestGRPCLayer) SpawnService(ctx context.Context, address string, register func(ctx context.Context, srv *grpc.Server) error, otherOptions ...grpc.ServerOption) *TestListener {
	listenerParts := strings.Split(address, ":")
	if len(listenerParts) < 1 {
		panic(fmt.Sprintf("GRPC listener has less than one part after split: %q", address))
	}
	hostName := listenerParts[0]

	tlsKeys, err := t.publicKeyInfra.NewKeyPair(hostName)
	require.NoError(t.t, err)

	serviceTLS, err := tlsKeys.ConfigureTLSConfig(&tls.Config{})
	require.NoError(t.t, err)

	t.updateResolver(ctx, address, hostName)

	opts := append(otherOptions, grpc.Creds(credentials.NewTLS(serviceTLS)))
	return &TestListener{
		address:  address,
		on:       t,
		register: register,
		opts:     opts,
	}
}

func (t *TestGRPCLayer) updateResolver(ctx context.Context, address, hostName string) {
	t.resolverControl.Lock()
	defer t.resolverControl.Unlock()

	t.resolverState.Addresses = append(t.resolverState.Addresses, resolver.Address{Addr: address, ServerName: hostName})
	t.resolver.InitialState(t.resolverState)
}

func (t *TestGRPCLayer) Connect(ctx context.Context, address string, opts ...grpc.DialOption) *grpc.ClientConn {
	grpcClientOpts := append(t.ConnectOptions(), opts...)
	conn, err := grpc.NewClient(address, grpcClientOpts...)
	require.NoError(t.t, err)
	return conn
}

func (t *TestGRPCLayer) ConnectOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithResolvers(t.resolver),
		grpc.WithContextDialer(t.physicalTransport.Dial),
		grpc.WithTransportCredentials(credentials.NewTLS(t.publicKeyInfra.GenerateTLSConfig())),
	}
}

type TestListener struct {
	address  string
	on       *TestGRPCLayer
	register func(ctx context.Context, srv *grpc.Server) error
	opts     []grpc.ServerOption
}

func (t *TestListener) Serve(ctx context.Context) error {
	port, err := t.on.physicalTransport.Listen(ctx, t.address)
	if err != nil {
		return err
	}

	srv := grpc.NewServer(t.opts...)
	if err := t.register(ctx, srv); err != nil {
		return err
	}

	listenerOut := make(chan error, 1)
	go func() {
		listenerOut <- srv.Serve(port)
	}()

	select {
	case err := <-listenerOut:
		return err
	case <-ctx.Done():
		//todo: configurable shutdown time?
		shutdownTimeout := time.NewTimer(100 * time.Millisecond)
		defer shutdownTimeout.Stop()

		srv.GracefulStop()
		select {
		case <-shutdownTimeout.C:
			srv.Stop()
			return errors.New("timed out waiting for listener to gracefully stop")
		case err := <-listenerOut:
			return err
		}
	}
}
