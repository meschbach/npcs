package competition

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"time"
)

type GRPCEventKind int

const (
	GRPCEventReady GRPCEventKind = iota
)

type GRPCService struct {
	network  Network
	address  string
	register func(ctx context.Context, srv *grpc.Server) error
	opts     []grpc.ServerOption
	events   chan GRPCEventKind
}

func (g *GRPCService) Serve(ctx context.Context) error {
	server := grpc.NewServer(g.opts...)
	if err := g.register(ctx, server); err != nil {
		return err
	}

	listener, err := g.network.Listen(ctx, g.address)
	if err != nil {
		return err
	}
	listenerErr := make(chan error, 1)
	(func() {
		if g.events != nil {
			g.events <- GRPCEventReady
		}
		fmt.Printf("grpc service at %q listening\n", g.address)
		err := server.Serve(listener)
		listenerErr <- err
	})()

	for {
		select {
		case problem := <-listenerErr:
			return problem
		case <-ctx.Done():
			var cleanUpErrors []error

			closeErr := listener.Close()
			cleanUpErrors = append(cleanUpErrors, closeErr)
			closeTimeout := time.NewTimer(1 * time.Second)
			select {
			case listenerErr <- closeErr:
				cleanUpErrors = append(cleanUpErrors, closeErr)
			case <-closeTimeout.C:
				cleanUpErrors = append(cleanUpErrors, errors.New(fmt.Sprintf("grpc service at %q failed to close", g.address)))
			}
			return errors.Join(cleanUpErrors...)
		}
	}
}
