package simple

import (
	"context"
	"github.com/meschbach/npcs/competition/wire"
	"github.com/meschbach/npcs/junk/realnet"
	"google.golang.org/grpc"
	"log/slog"
)

type RunOnceInstanceOpt func(*RunOnceInstance)

type RunOnceInstance struct {
	net realnet.GRPCNetwork

	matcherAddress string
	matcherOptions []grpc.DialOption

	instanceServiceAddress string
	instanceServiceOptions []grpc.DialOption
}

func WithInstanceNetwork(net realnet.GRPCNetwork) RunOnceInstanceOpt {
	return func(o *RunOnceInstance) { o.net = net }
}

func WithInstanceMatcherAddress(address string) RunOnceInstanceOpt {
	return func(o *RunOnceInstance) { o.matcherAddress = address }
}

func WithInstanceAddress(address string) RunOnceInstanceOpt {
	return func(o *RunOnceInstance) { o.instanceServiceAddress = address }
}

func NewRunOnceInstance(opts ...RunOnceInstanceOpt) *RunOnceInstance {
	r := &RunOnceInstance{}
	for _, o := range opts {
		o(r)
	}
	return r
}

func (r *RunOnceInstance) Run(ctx context.Context) error {
	// build service
	game := NewGameService()

	// export service
	service, err := r.net.Listener(ctx, r.instanceServiceAddress, func(ctx context.Context, server *grpc.Server) error {
		RegisterSimpleGameServer(server, game)
		return nil
	})
	if err != nil {
		return err
	}
	if err := service.Start(ctx); err != nil {
		return err
	}

	// register with service
	slog.InfoContext(ctx, "registering game")
	matcherConnection, err := r.net.Client(ctx, r.matcherAddress, r.matcherOptions...)
	if err != nil {
		return err
	}
	registryClient := wire.NewGameRegistryClient(matcherConnection)
	if _, err := registryClient.RegisterGame(ctx, &wire.RegisterGameIn{
		Name: "github.com/meschbach/npc/competition/simple/v0",
	}); err != nil {
		return err
	}
	engineOrchestrationClient := wire.NewGameEngineOrchestrationClient(matcherConnection)
	if _, err := engineOrchestrationClient.EngineAvailable(ctx, &wire.EngineAvailableIn{
		ForGame:  "github.com/meschbach/npc/competition/simple/v0",
		StartURL: r.instanceServiceAddress,
	}); err != nil {
		return err
	}
	slog.InfoContext(ctx, "simple game engine started")
	return nil
}
