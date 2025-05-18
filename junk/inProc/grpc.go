package inProc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/madflojo/testcerts"
	"github.com/meschbach/npcs/junk"
	"github.com/meschbach/npcs/junk/realnet"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
	"log/slog"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

type GRPCNetwork struct {
	t                 *testing.T
	physicalTransport *Network
	publicKeyInfra    *testcerts.CertificateAuthority

	//todo: resolver is a mess in terms of initial seeds versus actual.  might be better to build out a custom binding
	//to avoid needing to update.
	resolverControl sync.Mutex
	resolverState   resolver.State
	resolver        *manual.Resolver
}

func NewGRPCNetwork(t *testing.T, opts ...GRPCOption) *GRPCNetwork {
	layer := &GRPCNetwork{
		t:                 t,
		publicKeyInfra:    testcerts.NewCA(),
		physicalTransport: NewNetwork(),
		resolverControl:   sync.Mutex{},
		resolverState:     resolver.State{},
	}
	for _, opt := range opts {
		opt(layer)
	}
	return layer
}

type GRPCOption func(*GRPCNetwork)

func WithNetwork(net *Network) GRPCOption {
	return func(test *GRPCNetwork) {
		test.physicalTransport = net
	}
}

func (t *GRPCNetwork) Listener(ctx context.Context, address string, registratar realnet.GRPCRegister, options ...grpc.ServerOption) (junk.Component, error) {
	return t.SpawnService(ctx, address, registratar, options...), nil
}

func (t *GRPCNetwork) Client(ctx context.Context, address string, options ...grpc.DialOption) (*grpc.ClientConn, error) {
	internalURL := "in-proc://" + address
	slog.InfoContext(ctx, "New GRPC client", "address", internalURL)
	return t.Connect(ctx, internalURL, options...), nil
}

func (t *GRPCNetwork) SpawnService(ctx context.Context, address string, register func(ctx context.Context, srv *grpc.Server) error, otherOptions ...grpc.ServerOption) *TestListener {
	listenerParts := strings.Split(address, ":")
	if len(listenerParts) < 1 {
		panic(fmt.Sprintf("GRPC listener has less than one part after split: %q", address))
	}
	hostName := listenerParts[0]

	tlsKeys, err := t.publicKeyInfra.NewKeyPair(hostName)
	require.NoError(t.t, err)

	serviceTLS, err := tlsKeys.ConfigureTLSConfig(&tls.Config{})
	require.NoError(t.t, err)

	opts := append(otherOptions, grpc.Creds(credentials.NewTLS(serviceTLS)))
	return &TestListener{
		address:  address,
		on:       t,
		register: register,
		opts:     opts,
	}
}

func (t *GRPCNetwork) Connect(ctx context.Context, address string, opts ...grpc.DialOption) *grpc.ClientConn {
	grpcClientOpts := append(t.ConnectOptions(), opts...)
	conn, err := grpc.NewClient(address, grpcClientOpts...)
	require.NoError(t.t, err)
	return conn
}

func (t *GRPCNetwork) ConnectOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithResolvers(&grpcResolverBuilder{state: &sync.Mutex{}}),
		grpc.WithContextDialer(t.physicalTransport.Dial),
		grpc.WithTransportCredentials(credentials.NewTLS(t.publicKeyInfra.GenerateTLSConfig())),
	}
}

type TestListener struct {
	address  string
	on       *GRPCNetwork
	register func(ctx context.Context, srv *grpc.Server) error
	opts     []grpc.ServerOption
	port     net.Listener
	service  *grpc.Server
}

func (t *TestListener) Start(ctx context.Context) error {
	port, err := t.on.physicalTransport.Listen(ctx, t.address)
	if err != nil {
		return err
	}
	t.port = port

	slog.InfoContext(ctx, "gRPC inproc service starting", "address.given", t.address, "address.resolved", port.Addr().String())
	srv := grpc.NewServer(t.opts...)
	if err := t.register(ctx, srv); err != nil {
		slog.ErrorContext(ctx, "gRPC server failed to register", "error", err)
		return err
	}
	t.service = srv

	go func() {
		if err := srv.Serve(port); err != nil {
			slog.ErrorContext(ctx, "gRPC server failed to serve", "error", err)
		}
	}()
	return nil
}

func (t *TestListener) Stop(shutdownCtx context.Context) error {
	if t.service != nil {
		t.service.Stop()
	}
	if t.port != nil {
		return t.port.Close()
	}
	return nil
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
