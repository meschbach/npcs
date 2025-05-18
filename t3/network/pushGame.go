package network

import (
	"context"
	"errors"
	"github.com/meschbach/npcs/t3"
	"google.golang.org/grpc"
	"io"
)

type PushClient struct {
	server string
	gameID string
	//todo: we aren't using this
	token  string
	player Session

	grpcOpts []grpc.DialOption

	playerID int64
}

func NewPushClient(server, gameID, token string, player Session, grpcOpts ...grpc.DialOption) *PushClient {
	return &PushClient{
		server: server,
		gameID: gameID,
		token:  token,
		player: player,

		grpcOpts: grpcOpts,
	}
}

func (p *PushClient) Serve(ctx context.Context) (problem error) {
	//dial out
	conn, err := grpc.NewClient(p.server, p.grpcOpts...)
	if err != nil {
		return err
	}
	defer func() {
		problem = errors.Join(problem, conn.Close())
	}()
	c := NewT3PushClient(conn)
	feed, err := c.ConnectToGame(ctx, &JoinGameIn{GameID: p.gameID})
	if err != nil {
		return err
	}
	for {
		m, err := feed.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if m.Initial != nil {
			p.playerID = m.Initial.YourPlayer
		}
		if m.Move != nil {
			if err := p.player.MoveMade(ctx, t3.Move{
				Player: int(m.Move.Player),
				Row:    int(m.Move.Row),
				Column: int(m.Move.Column),
			}); err != nil {
				return err
			}
		}
		if m.DoTurn != nil {
			if err := p.pushMove(ctx, c); err != nil {
				return err
			}
		}

		if m.Conclusion != nil {
			if err := p.player.Done(ctx); err != nil {
				return err
			}
			return nil
		}
	}
}

func (p *PushClient) pushMove(ctx context.Context, c T3PushClient) error {
	move, err := p.player.NextMove(ctx)
	if err != nil {
		return err
	}
	_, err = c.Move(ctx, &PushMoveIn{
		GameID: p.gameID,
		Move: &NextMoveOut{
			Row:    int64(move.Row),
			Column: int64(move.Column),
		},
	})
	return err
}

func (p *PushClient) doMove(ctx context.Context, client T3PushClient) error {
	move, err := p.player.NextMove(ctx)
	if err != nil {
		return err
	}
	if _, err := client.Move(ctx, &PushMoveIn{
		GameID: p.gameID,
		Move:   &NextMoveOut{Row: int64(move.Row), Column: int64(move.Column)},
	}); err != nil {
		return err
	}
	return nil
}
