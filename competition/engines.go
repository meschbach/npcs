package competition

import (
	"context"
	"github.com/meschbach/npcs/competition/wire"
)

type gameEngine struct {
	id        string
	engineURL string
	started   bool
}

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
	id, err := e.core.registerInstance(ctx, in.ForGame, in.StartURL)
	if err != nil {
		return nil, err
	}

	return &wire.EngineAvailableOut{GameID: id}, nil
}
