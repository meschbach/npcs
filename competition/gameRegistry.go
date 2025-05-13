package competition

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/meschbach/npcs/competition/wire"
	"sync"
)

type registeredGame struct {
	id string
}

type GameRegistryService struct {
	wire.UnimplementedGameRegistryServer
	state *sync.Mutex
	games map[string]registeredGame
}

func NewGameRegistryService() *GameRegistryService {
	return &GameRegistryService{
		games: make(map[string]registeredGame),
		state: &sync.Mutex{},
	}
}

func (g *GameRegistryService) ListRegisteredGames(ctx context.Context, in *wire.ListRegisteredGamesIn) (*wire.ListRegisteredGamesOut, error) {
	g.state.Lock()
	defer g.state.Unlock()
	out := &wire.ListRegisteredGamesOut{}
	for name, game := range g.games {
		out.Games = append(out.Games, &wire.RegisteredGame{
			Name: name,
			Id:   game.id,
		})
	}
	return out, nil
}

func (g *GameRegistryService) RegisterGame(ctx context.Context, in *wire.RegisterGameIn) (*wire.RegisterGameOut, error) {
	g.state.Lock()
	defer g.state.Unlock()
	if _, exists := g.games[in.Name]; exists {
		return nil, errors.New("game already registered")
	}
	structuredID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	id := structuredID.String()
	g.games[in.Name] = registeredGame{id: id}
	return &wire.RegisterGameOut{
		Id: id,
	}, nil
}

func (g *GameRegistryService) findGame(name string) (bool, error) {
	g.state.Lock()
	defer g.state.Unlock()
	_, has := g.games[name]
	return has, nil
}
