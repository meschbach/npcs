package t3

import (
	"context"
	"errors"
)

var PlayerDisconnected = errors.New("player disconnected")

type Player interface {
	NextPlay(ctx context.Context) (Move, error)
	PushHistory(ctx context.Context, move Move) error
}

type ChanPlayer struct {
	input  <-chan Move
	output chan<- Move
}

func NewPlayer(input <-chan Move) *ChanPlayer {
	//outputs := make(chan Move, 8)
	return &ChanPlayer{
		input: input,
		//output: outputs,
	}
}

func (t *ChanPlayer) NextPlay(ctx context.Context) (Move, error) {
	select {
	case <-ctx.Done():
		return Move{}, errors.Join(ctx.Err(), PlayerDisconnected)
	case move := <-t.input:
		return move, nil
	}
}

func (t *ChanPlayer) PushHistory(ctx context.Context, move Move) error {
	if t.output == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return errors.Join(ctx.Err(), PlayerDisconnected)
	case t.output <- move:
		return nil
	}
}
