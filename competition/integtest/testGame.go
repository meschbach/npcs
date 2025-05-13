package integtest

import (
	"context"
	"github.com/meschbach/npcs/competition/wire"
	"github.com/meschbach/npcs/junk/inProc"
	"google.golang.org/grpc"
	"sync"
)

// TestGameEngine is a simple service to fake games with pre-determined winners.
type TestGameEngine struct {
	wire.UnimplementedSimpleTestGameServiceServer

	gameName             string
	orchestrationAddress string
	bindTo               string
	layer                *inProc.TestGRPCLayer

	state             *sync.Mutex
	gameID            *string
	gameDoneCondition *sync.Cond
	//game ids connected
	connections []string
}

func newTestGameEngine(name, orchestrationAddress, bindTo string, network *inProc.TestGRPCLayer) *TestGameEngine {
	stateLock := &sync.Mutex{}
	return &TestGameEngine{
		gameName:             name,
		orchestrationAddress: orchestrationAddress,
		bindTo:               bindTo,
		layer:                network,
		state:                stateLock,
		gameID:               nil,
		gameDoneCondition:    sync.NewCond(stateLock),
	}
}

func (t *TestGameEngine) Connected(ctx context.Context, in *wire.SimpleTestGameIn) (*wire.SimpleTestGameOut, error) {
	t.state.Lock()
	defer t.state.Unlock()

	t.connections = append(t.connections, in.GameID)
	return &wire.SimpleTestGameOut{}, nil
}

func (t *TestGameEngine) Serve(ctx context.Context) error {
	l := t.layer.SpawnService(ctx, t.bindTo, func(ctx context.Context, srv *grpc.Server) error {
		wire.RegisterSimpleTestGameServiceServer(srv, t)
		return nil
	})
	listenerResult := make(chan error, 1)
	go func() {
		listenerResult <- l.Serve(ctx)
	}()

	orchestrationConnection := t.layer.Connect(ctx, t.orchestrationAddress)
	orchestrationClient := wire.NewGameEngineOrchestrationClient(orchestrationConnection)
	//todo: error handling
	defer orchestrationConnection.Close()
	reservation, err := orchestrationClient.EngineAvailable(ctx, &wire.EngineAvailableIn{
		ForGame:  t.gameName,
		StartURL: t.bindTo,
	})
	if err != nil {
		return err
	}

	t.state.Lock()
	t.gameID = &reservation.GameID
	t.gameDoneCondition.Wait()
	t.state.Unlock()

	_, err = orchestrationClient.GameComplete(ctx, &wire.EngineGameCompleteIn{
		GameID: reservation.GameID,
	})
	return err
}
