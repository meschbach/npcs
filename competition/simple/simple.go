// Package simple provides a simple game service with a pre-determined winner and completion once a certain threshold
// of players join.  Intended as a simple testing service.
package simple

import (
	"context"
	"log/slog"
	"sync"
)

type GameService struct {
	UnimplementedSimpleGameServer
	state        *sync.Mutex
	othersJoined *sync.Cond
	joinedCount  int
}

func NewGameService() *GameService {
	s := &sync.Mutex{}
	return &GameService{
		state:        s,
		othersJoined: sync.NewCond(s),
		joinedCount:  0,
	}
}

func (s *GameService) Joined(ctx context.Context, j *JoinedIn) (*JoinedOut, error) {
	slog.InfoContext(ctx, "Player joining")

	s.state.Lock()
	defer s.state.Unlock()

	won := s.joinedCount == 0
	s.joinedCount++
	for s.joinedCount < 2 {
		slog.InfoContext(ctx, "Awaiting additional players to join...")
		//todo: handle context timeout
		s.othersJoined.Wait()
	}
	slog.InfoContext(ctx, "Players joined", "won", won)
	s.othersJoined.Signal()

	return &JoinedOut{
		Won: won,
	}, nil
}
