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

type remotePlayer struct {
	wire   network.T3Client
	gameID int64
}

func (r *remotePlayer) NextPlay(ctx context.Context) (t3.Move, error) {
	reply, err := r.wire.NextMove(ctx, &network.NextMoveIn{
		GameID: r.gameID,
	})
	if err != nil {
		return t3.Move{}, err
	}
	return t3.Move{
		Row:    int(reply.Row),
		Column: int(reply.Column),
	}, nil
}

func (r *remotePlayer) PushHistory(ctx context.Context, move t3.Move) error {
	_, err := r.wire.MoveMade(ctx, &network.MoveMadeIn{
		Player: int64(move.Player),
		Row:    int64(move.Row),
		Column: int64(move.Column),
		GameID: r.gameID,
	})
	return err
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
		nextPlayer, players = players[0], players[1:]
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
	firstPlayerStart, err := client.StartGame(ctx, &network.StartGameIn{YourPlayer: 1})
	require.NoError(t, err)
	assert.NotNil(t, firstPlayerStart)

	secondPlayerStart, err := client.StartGame(ctx, &network.StartGameIn{YourPlayer: 2})
	require.NoError(t, err)
	assert.NotNil(t, secondPlayerStart)

	game := t3.NewGame(&remotePlayer{
		wire:   client,
		gameID: firstPlayerStart.GameID,
	}, &remotePlayer{
		wire:   client,
		gameID: secondPlayerStart.GameID,
	})
	for !game.Concluded() {
		require.NoError(t, game.Step(ctx))
	}
	assert.True(t, game.Concluded())
	finished, winner := game.Result()
	assert.True(t, finished)
	assert.Equal(t, 1, winner)
}
