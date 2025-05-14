package competition

import (
	"context"
	"github.com/meschbach/npcs/competition/wire"
)

type gameRegistryService struct {
	wire.UnimplementedGameRegistryServer
	core *matcher
}

func newGameRegistryService(core *matcher) *gameRegistryService {
	return &gameRegistryService{
		core: core,
	}
}

func (g *gameRegistryService) ListRegisteredGames(ctx context.Context, in *wire.ListRegisteredGamesIn) (*wire.ListRegisteredGamesOut, error) {
	g.core.state.Lock()
	defer g.core.state.Unlock()

	out := &wire.ListRegisteredGamesOut{}
	for name, _ := range g.core.gameMatches {
		out.Games = append(out.Games, &wire.RegisteredGame{
			Name: name,
			Id:   name,
		})
	}
	return out, nil
}

func (g *gameRegistryService) RegisterGame(ctx context.Context, in *wire.RegisterGameIn) (*wire.RegisterGameOut, error) {
	err := g.core.ensureGame(ctx, in.Name)
	return &wire.RegisterGameOut{
		Id: in.Name,
	}, err
}
