package competition

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/meschbach/npcs/competition/wire"
	"sync"
)

type gameEngine struct {
	id        string
	engineURL string
	started   bool
}

type engines struct {
	wire.UnimplementedGameEngineOrchestrationServer
	state    *sync.Mutex
	registry *GameRegistryService
	//game name -> game id -> instance
	gameEngines map[string]map[string]*gameEngine
}

func newEngines(registry *GameRegistryService) *engines {
	return &engines{
		registry:    registry,
		state:       &sync.Mutex{},
		gameEngines: make(map[string]map[string]*gameEngine),
	}
}

func (e *engines) EngineAvailable(ctx context.Context, in *wire.EngineAvailableIn) (*wire.EngineAvailableOut, error) {
	has, err := e.registry.findGame(in.ForGame)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("game not registered")
	}

	structuredID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	id := structuredID.String()

	e.state.Lock()
	defer e.state.Unlock()
	if _, has := e.gameEngines[in.ForGame]; !has {
		e.gameEngines[in.ForGame] = make(map[string]*gameEngine)
	}
	e.gameEngines[in.ForGame][id] = &gameEngine{
		id:        id,
		engineURL: in.StartURL,
		started:   false,
	}
	return &wire.EngineAvailableOut{GameID: id}, nil
}
