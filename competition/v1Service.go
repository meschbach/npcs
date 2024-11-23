package competition

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/meschbach/npcs/competition/wire"
	"sync"
)

type v1Service struct {
	wire.CompetitionV1Server
	auth         Auth
	t3MatchesURL string

	lock       sync.Mutex
	persistent map[string]*persistentPlayer
}

func (v *v1Service) RegisterPersistentPlayer(ctx context.Context, in *wire.RegisterPlayerIn) (*wire.RegisterPlayerOut, error) {
	userID, err := v.auth.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	var player *persistentPlayer
	(func() {
		v.lock.Lock()
		defer v.lock.Unlock()

		if p, has := v.persistent[userID]; has {
			player = p
		} else {
			player = &persistentPlayer{
				lock:   sync.Mutex{},
				agents: make(map[string]*persistentAgent),
			}
			v.persistent[userID] = player
		}
	})()

	player.lock.Lock()
	defer player.lock.Unlock()

	if _, has := player.agents[in.Name]; has {
		return nil, errors.New("agent already exists")
	}
	player.agents[in.Name] = &persistentAgent{
		agentURL: in.OrchestrationURL,
	}
	return &wire.RegisterPlayerOut{}, nil
}

func (v *v1Service) QuickMatch(ctx context.Context, in *wire.QuickMatchIn) (*wire.QuickMatchOut, error) {
	v.lock.Lock()
	defer v.lock.Unlock()

	//new game
	gameID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	//find an available player
	var agent *persistentAgent
	for _, player := range v.persistent {
		agent = player.findAvailableAgent(gameID)
		if agent != nil {
			break
		}
	}
	if agent == nil {
		return nil, errors.New("no agents available agents")
	}

	//spawn match controller
	//todo: consider architecture for horizontal scalability here

	//spawn persistent game bridge
	//spawn quick game bridge

	//notify the user for connection
	return &wire.QuickMatchOut{
		MatchURL: v.t3MatchesURL,
		UUID:     gameID.String(),
	}, nil
}
