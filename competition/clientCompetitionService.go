package competition

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/meschbach/npcs/competition/wire"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
)

type completedGameResult struct {
	winner string
}

// clientCompetitionService is a gRPC server implementation for managing player competitions and game results.
// It embeds wire.UnimplementedCompetitionV1Server to ensure forward compatibility with the CompetitionV1 interface.
// auth is used to authenticate and retrieve the user information for managing persistent players.
// t3MatchesURL is the base URL for providing match connection details in QuickMatch responses.
// lock is a mutex to safely manage concurrent access to the service's state.
// persistent is a map storing players and their associated agents for persistent gameplay sessions.
// gameResult is a map tracking completed game results associated by game IDs.
type clientCompetitionService struct {
	wire.UnimplementedCompetitionV1Server
	auth         Auth
	t3MatchesURL string

	lock       sync.Mutex
	persistent map[string]*persistentPlayer
	gameResult map[string]*completedGameResult
}

func (v *clientCompetitionService) RegisterPersistentPlayer(ctx context.Context, in *wire.RegisterPlayerIn) (*wire.RegisterPlayerOut, error) {
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

	//todo: move into player?
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

func (v *clientCompetitionService) QuickMatch(ctx context.Context, in *wire.QuickMatchIn) (*wire.QuickMatchOut, error) {
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
		return nil, status.Error(codes.NotFound, "agent not found")
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

func (v *clientCompetitionService) GameResult(ctx context.Context, in *wire.GameResultIn) (*wire.GameResultOut, error) {
	v.lock.Lock()
	defer v.lock.Unlock()

	if game, ok := v.gameResult[in.GameID]; ok {
		return &wire.GameResultOut{
			WinningPlayerName: game.winner,
		}, nil
	} else {
		return nil, status.Error(codes.NotFound, "game not found")
	}
}
