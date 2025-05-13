package competition

import (
	"context"
	"net"
)

type Network interface {
	Listen(ctx context.Context, address string) (net.Listener, error)
	Dial(ctx context.Context, address string) (net.Conn, error)
}

type Component interface {
	Start(ctx context.Context) error
	Stop(shutdownCtx context.Context) error
}
