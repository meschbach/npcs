package network

import (
	"context"
	"errors"
	"io"

	"github.com/meschbach/npcs/t3"
	"google.golang.org/grpc"
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
		done, err := p.handleMessage(ctx, c, m)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}
}

func (p *PushClient) handleMessage(ctx context.Context, c T3PushClient, m *T3PushEvent) (bool, error) {
	p.handleInitial(m)
	if err := p.handleMove(ctx, m); err != nil {
		return false, err
	}
	if err := p.handleDoTurn(ctx, c, m); err != nil {
		return false, err
	}
	return p.handleConclusion(ctx, m)
}

func (p *PushClient) handleInitial(m *T3PushEvent) {
	if m.Initial != nil {
		p.playerID = m.Initial.YourPlayer
	}
}

func (p *PushClient) handleMove(ctx context.Context, m *T3PushEvent) error {
	if m.Move == nil {
		return nil
	}
	return p.player.MoveMade(ctx, t3.Move{
		Player: int(m.Move.Player),
		Row:    int(m.Move.Row),
		Column: int(m.Move.Column),
	})
}

func (p *PushClient) handleDoTurn(ctx context.Context, c T3PushClient, m *T3PushEvent) error {
	if m.DoTurn == nil {
		return nil
	}
	return p.pushMove(ctx, c)
}

func (p *PushClient) handleConclusion(ctx context.Context, m *T3PushEvent) (bool, error) {
	if m.Conclusion == nil {
		return false, nil
	}
	if err := p.player.Done(ctx); err != nil {
		return false, err
	}
	return true, nil
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
