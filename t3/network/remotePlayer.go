package network

import (
	"context"
	"github.com/meschbach/npcs/t3"
)

func NewRemotePlayer(ctx context.Context, client T3Client, playerID int64) (*RemotePlayer, error) {
	reply, err := client.StartGame(ctx, &StartGameIn{
		YourPlayer: playerID,
	})
	if err != nil {
		return nil, err
	}
	return &RemotePlayer{
		wire:   client,
		gameID: reply.GameID,
	}, nil
}

type RemotePlayer struct {
	wire   T3Client
	gameID int64
}

func (r *RemotePlayer) NextPlay(ctx context.Context) (t3.Move, error) {
	reply, err := r.wire.NextMove(ctx, &NextMoveIn{
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

func (r *RemotePlayer) PushHistory(ctx context.Context, move t3.Move) error {
	_, err := r.wire.MoveMade(ctx, &MoveMadeIn{
		Player: int64(move.Player),
		Row:    int64(move.Row),
		Column: int64(move.Column),
		GameID: r.gameID,
	})
	return err
}
