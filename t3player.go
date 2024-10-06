package npcs

import (
	"context"
	"errors"
)

var PlayerDisconnected = errors.New("player disconnected")

type T3Player struct {
	input <-chan T3Move
	//I got ahead of myself
	//output chan<- T3Move
}

func NewT3Player(input <-chan T3Move) *T3Player {
	//outputs := make(chan T3Move, 8)
	return &T3Player{
		input: input,
		//output: outputs,
	}
}

func (t *T3Player) NextPlay(ctx context.Context) (T3Move, error) {
	select {
	case <-ctx.Done():
		return T3Move{}, errors.Join(ctx.Err(), PlayerDisconnected)
	case move := <-t.input:
		return move, nil
	}
}

//func (t *T3Player) PushHistory(ctx context.Context, move T3Move) error {
//	select {
//	case <-ctx.Done():
//		return errors.Join(ctx.Err(), PlayerDisconnected)
//	case t.output <- move:
//		return nil
//	}
//}
