package realnet

import (
	"context"
	junk2 "github.com/meschbach/npcs/junk"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type GRPCRegister func(ctx context.Context, server *grpc.Server) error

// GRPCNetwork is a seam for testing.
type GRPCNetwork interface {
	Listener(ctx context.Context, address string, register GRPCRegister, options ...grpc.ServerOption) (junk2.Component, error)
	Client(ctx context.Context, address string, options ...grpc.DialOption) (*grpc.ClientConn, error)
}

// RealGRPCNetwork is a concrete implementation for managing gRPC network listeners and clients.
type RealGRPCNetwork struct{}

func (r *RealGRPCNetwork) Listener(ctx context.Context, address string, registrar GRPCRegister, options ...grpc.ServerOption) (junk2.Component, error) {
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
	registrar GRPCRegister
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

	slog.InfoContext(ctx, "gRPC service starting\n", "address.given", g.address, "address.resolved", lis.Addr().String())
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
