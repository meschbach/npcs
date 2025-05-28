package realnet

import (
	"context"
	"fmt"
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

type ClientConnectError struct {
	address    string
	underlying error
}

func (c *ClientConnectError) Unwrap() error {
	return c.underlying
}

func (c *ClientConnectError) Error() string {
	return fmt.Sprintf("failed to connect to gRPC service at %s: %s", c.address, c.underlying.Error())
}

func (r *RealGRPCNetwork) Client(ctx context.Context, address string, options ...grpc.DialOption) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(address, options...)
	if err != nil {
		err = &ClientConnectError{
			address:    address,
			underlying: err,
		}
	}
	return conn, err
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
