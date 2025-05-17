// Package simple provides a simple game service with a pre-determined winner and completion once a certain threshold
// of players join.  Intended as a simple testing service.
package simple

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"log/slog"
	"sync"
)

type gamePhase int

const (
	gamePhase_waitingForPlayers gamePhase = iota
	gamePhase_done
)

type gameInstance struct {
	state   *sync.Mutex
	changes *sync.Cond

	phase       gamePhase
	joinedCount int
	players     []string
	winner      string
}

func (g *gameInstance) joinPlayer(ctx context.Context, player string) bool {
	g.state.Lock()
	defer g.state.Unlock()

	won := g.joinedCount == 0
	g.players = append(g.players, player)
	g.joinedCount++
	for g.joinedCount < 2 {
		slog.InfoContext(ctx, "Awaiting additional players to join...")
		//todo: handle context timeout
		g.changes.Wait()
	}
	if won {
		g.winner = player
	}
	if g.phase != gamePhase_done {
		g.phase = gamePhase_done
	}

	slog.InfoContext(ctx, "Players joined", "won", won)
	g.changes.Broadcast()

	return won
}

func (g *gameInstance) waitOnGameCompletion() {
	slog.Info("gameInstance waiting on completion")
	g.state.Lock()
	defer g.state.Unlock()

	for g.phase != gamePhase_done {
		slog.Info("gameInstance waiting")
		g.changes.Wait()
	}
	slog.Info("gameInstance game completed.")
}

type GameService struct {
	UnimplementedSimpleGameServer
	state        *sync.Mutex
	othersJoined *sync.Cond
	instances    map[string]*gameInstance
}

func NewGameService() *GameService {
	s := &sync.Mutex{}
	return &GameService{
		state:        s,
		othersJoined: sync.NewCond(s),
		instances:    make(map[string]*gameInstance),
	}
}

func (s *GameService) findInstance(instance string) (*gameInstance, bool) {
	s.state.Lock()
	defer s.state.Unlock()

	gameInstance, has := s.instances[instance]
	return gameInstance, has
}

func (s *GameService) Joined(ctx context.Context, j *JoinedIn) (*JoinedOut, error) {
	slog.InfoContext(ctx, "Player joining", "instance", j.InstanceID)

	instance, has := s.findInstance(j.InstanceID)
	if !has {
		return nil, errors.New("instance not found")
	}
	won := instance.joinPlayer(ctx, j.InstanceID)

	return &JoinedOut{
		Won: won,
	}, nil
}

func (s *GameService) runGameInstance() (string, *gameInstance, error) {
	instanceIDStruct, err := uuid.NewV7()
	if err != nil {
		return "", nil, err
	}
	instanceID := instanceIDStruct.String()

	lock := &sync.Mutex{}
	g := &gameInstance{
		state:       lock,
		changes:     sync.NewCond(lock),
		phase:       gamePhase_waitingForPlayers,
		joinedCount: 0,
	}

	s.addSession(instanceID, g)
	return instanceID, g, nil
}

func (s *GameService) addSession(id string, g *gameInstance) {
	slog.Info("adding session", "instance", id)
	s.state.Lock()
	defer s.state.Unlock()
	s.instances[id] = g
}
