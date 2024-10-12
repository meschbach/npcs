package systest

import (
	"context"
	"errors"
	"github.com/meschbach/npcs/t3"
	"github.com/meschbach/npcs/t3/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"sync"
	"testing"
	"time"
)

type ScriptedPlayer struct {
	moves []t3.Move
}

func (s *ScriptedPlayer) MoveMade(ctx context.Context, otherPlayer t3.Move) error {
	return nil
}

func (s *ScriptedPlayer) NextMove(ctx context.Context) (move t3.Move, err error) {
	count := len(s.moves)
	if count == 0 {
		return move, errors.New("no moves left")
	} else if count == 1 {
		move = s.moves[0]
		s.moves = nil
	} else {
		move, s.moves = s.moves[0], s.moves[1:]
	}
	return move, nil
}

func (s *ScriptedPlayer) Done(ctx context.Context) error {
	return nil
}

func TestNetworkedGame(t *testing.T) {
	ctx, done := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(done)

	dequeueLock := &sync.Mutex{}
	players := []network.Session{
		&ScriptedPlayer{moves: []t3.Move{
			{Row: 0, Column: 0},
			{Row: 0, Column: 1},
			{Row: 0, Column: 2},
		}},
		&ScriptedPlayer{moves: []t3.Move{
			{Row: 2, Column: 2},
			{Row: 2, Column: 1},
			{Row: 2, Column: 0},
		}},
	}
	h := network.NewHub(func(ctx context.Context) (network.Session, error) {
		dequeueLock.Lock()
		defer dequeueLock.Unlock()

		var nextPlayer network.Session
		nextPlayer, players = unshiftSlice(players)
		return nextPlayer, nil
	})

	fabric := bufconn.Listen(32 * 1024)
	s := grpc.NewServer()
	network.RegisterT3Server(s, h)
	go func() {
		require.NoError(t, s.Serve(fabric))
	}()
	t.Cleanup(func() {
		s.GracefulStop()
	})

	c, err := grpc.DialContext(ctx, "bufnet", grpc.WithInsecure(), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return fabric.DialContext(ctx)
	}))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, c.Close())
	})

	client := network.NewT3Client(c)

	p1, err := network.NewRemotePlayer(ctx, client, 1)
	require.NoError(t, err)
	p2, err := network.NewRemotePlayer(ctx, client, 2)
	require.NoError(t, err)

	game := t3.NewGame(p1, p2)
	for !game.Concluded() {
		require.NoError(t, game.Step(ctx))
	}
	assert.True(t, game.Concluded())
	finished, winner := game.Result()
	assert.True(t, finished)
	assert.Equal(t, 1, winner)
}

func unshiftSlice[T any](in []T) (out T, remainder []T) {
	count := len(in)
	if count == 0 {
		return out, remainder
	} else if count == 1 {
		return in[0], nil
	} else {
		return in[0], in[1:]
	}
}
