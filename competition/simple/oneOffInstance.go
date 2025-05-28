package simple

import (
	"context"
	"github.com/meschbach/npcs/competition/wire"
	"github.com/meschbach/npcs/junk/realnet"
	"google.golang.org/grpc"
	"log/slog"
	"sync"
)

type RunOnceInstancePhase int

const (
	RunOnceInstancePhase_init RunOnceInstancePhase = iota
	RunOnceInstancePhase_started
	RunOnceInstancePhase_done
)

type RunOnceInstanceOpt func(*RunOnceInstance)

type RunOnceInstance struct {
	state   *sync.Mutex
	waiters *sync.Cond
	net     realnet.GRPCNetwork

	matcherAddress string
	matcherOptions []grpc.DialOption

	orchestration wire.GameEngineOrchestrationClient

	instanceServiceAddress string
	instanceServiceOptions []grpc.DialOption
	phase                  RunOnceInstancePhase
	instance               *GameInstance
	instanceID             string
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
	gate := &sync.Mutex{}
	r := &RunOnceInstance{
		state:   gate,
		waiters: sync.NewCond(gate),
		phase:   RunOnceInstancePhase_init,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

func (r *RunOnceInstance) Run(ctx context.Context) error {
	r.state.Lock()
	defer r.state.Unlock()

	// build service
	game := NewGameService()

	// export service
	service, err := r.net.Listener(ctx, r.instanceServiceAddress, func(ctx context.Context, server *grpc.Server) error {
		slog.InfoContext(ctx, "RunOnceInstance.Run.register")
		RegisterSimpleGameServer(server, game)
		return nil
	})
	if err != nil {
		return err
	}
	if err := service.Start(ctx); err != nil {
		return err
	}

	//
	gameID, gameInstance, err := game.RunGameInstance()
	if err != nil {
		return err
	}
	r.instance = gameInstance
	r.instanceID = gameID

	// register with service
	slog.InfoContext(ctx, "registering game", "instanceID", r.instanceID)
	matcherConnection, err := r.net.Client(ctx, r.matcherAddress, r.matcherOptions...)
	if err != nil {
		return err
	}
	registryClient := wire.NewGameRegistryClient(matcherConnection)
	_, err = registryClient.RegisterGame(ctx, &wire.RegisterGameIn{
		Name:       "github.com/meschbach/npc/competition/simple/v0",
		InstanceID: r.instanceID,
	})
	if err != nil {
		return err
	}
	engineOrchestrationClient := wire.NewGameEngineOrchestrationClient(matcherConnection)
	if _, err := engineOrchestrationClient.EngineAvailable(ctx, &wire.EngineAvailableIn{
		ForGame:    "github.com/meschbach/npc/competition/simple/v0",
		StartURL:   r.instanceServiceAddress,
		InstanceID: gameID,
	}); err != nil {
		return err
	}
	r.orchestration = engineOrchestrationClient
	slog.InfoContext(ctx, "simple game engine started")
	r.phase = RunOnceInstancePhase_started
	r.waiters.Broadcast()
	return nil
}

func (r *RunOnceInstance) WaitForStartup() {
	r.state.Lock()
	defer r.state.Unlock()
	for r.phase != RunOnceInstancePhase_started {
		r.waiters.Wait()
	}
}

func (r *RunOnceInstance) WaitForCompletion(ctx context.Context) error {
	slog.InfoContext(ctx, "RunOnceInstance.WaitForCompletion")
	done := make(chan int)
	go func() {
		r.instance.waitOnGameCompletion()
		done <- 0
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}
	slog.InfoContext(ctx, "Game completed.")

	_, err := r.orchestration.GameComplete(ctx, &wire.EngineGameCompleteIn{
		Results: &wire.CompletedGame{
			Game:       "github.com/meschbach/npc/competition/simple/v0",
			InstanceID: r.instanceID,
			Players:    r.instance.players,
			Winner:     r.instance.winner,
		},
	})

	r.state.Lock()
	defer r.state.Unlock()

	r.phase = RunOnceInstancePhase_done
	return err
}
