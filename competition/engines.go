package competition

import (
	"context"
	"errors"
	"github.com/meschbach/npcs/competition/wire"
)

type engines struct {
	wire.UnimplementedGameEngineOrchestrationServer
	core *matcher
}

func newEngines(core *matcher) *engines {
	return &engines{
		core: core,
	}
}

func (e *engines) EngineAvailable(ctx context.Context, in *wire.EngineAvailableIn) (*wire.EngineAvailableOut, error) {
	if in.InstanceID == "" {
		return nil, errors.New("instance ID must be provided")
	}
	id, err := e.core.registerInstance(ctx, in.ForGame, in.InstanceID, in.StartURL)
	if err != nil {
		return nil, err
	}

	return &wire.EngineAvailableOut{GameID: id}, nil
}

func (e *engines) GameComplete(ctx context.Context, in *wire.EngineGameCompleteIn) (*wire.EngineGameCompleteOut, error) {
	err := e.core.gameCompleted(ctx, in.Results.Game, in.Results.InstanceID, in.Results.Winner)
	return &wire.EngineGameCompleteOut{}, err
}
