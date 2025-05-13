package competition

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"net"
)

type GRPCRegistratar func(ctx context.Context, server *grpc.Server) error

// GRPCNetwork is a seam for testing.
type GRPCNetwork interface {
	Listener(ctx context.Context, address string, registratar GRPCRegistratar, options ...grpc.ServerOption) (Component, error)
	Client(ctx context.Context, address string, options ...grpc.DialOption) (*grpc.ClientConn, error)
}

// RealGRPCNetwork is a concrete implementation for managing gRPC network listeners and clients.
type RealGRPCNetwork struct{}

func (r *RealGRPCNetwork) Listener(ctx context.Context, address string, registrar GRPCRegistratar, options ...grpc.ServerOption) (Component, error) {
	return &gRPCListener{
		address:   address,
		registrar: registrar,
		options:   options,
	}, nil
}

func (r *RealGRPCNetwork) Client(ctx context.Context, address string, options ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.NewClient(address, options...)
}

type gRPCListener struct {
	address   string
	registrar GRPCRegistratar
	options   []grpc.ServerOption
	server    *grpc.Server
	listener  net.Listener
}

func (g *gRPCListener) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", g.address)
	if err != nil {
		return err
	}
	g.listener = lis

	fmt.Printf("Listening at %s (resvoled %q)\n", g.address, lis.Addr().String())
	g.server = grpc.NewServer(g.options...)
	if err := g.registrar(ctx, g.server); err != nil {
		return err
	}

	go g.server.Serve(lis)
	return nil
}

func (g *gRPCListener) Stop(shutdownCtx context.Context) error {
	if g.server != nil {
		g.server.GracefulStop()
	}
	if g.listener != nil {
		return g.listener.Close()
	}
	return nil
}

var NetworkedGRPC *RealGRPCNetwork = nil
