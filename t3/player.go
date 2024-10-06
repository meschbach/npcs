package t3

import (
	"context"
	"errors"
)

var PlayerDisconnected = errors.New("player disconnected")

type Player struct {
	input <-chan Move
	//I got ahead of myself
	//output chan<- Move
}

func NewPlayer(input <-chan Move) *Player {
	//outputs := make(chan Move, 8)
	return &Player{
		input: input,
		//output: outputs,
	}
}

func (t *Player) NextPlay(ctx context.Context) (Move, error) {
	select {
	case <-ctx.Done():
		return Move{}, errors.Join(ctx.Err(), PlayerDisconnected)
	case move := <-t.input:
		return move, nil
	}
}

//func (t *Player) PushHistory(ctx context.Context, move Move) error {
//	select {
//	case <-ctx.Done():
//		return errors.Join(ctx.Err(), PlayerDisconnected)
//	case t.output <- move:
//		return nil
//	}
//}
