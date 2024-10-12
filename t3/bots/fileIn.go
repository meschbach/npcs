package bots

import (
	"context"
	"errors"
	"github.com/meschbach/npcs/t3"
)

type FillInBotSession struct {
	nextPosition  t3.Move
	internalBoard *t3.Board
}

func (f *FillInBotSession) MoveMade(ctx context.Context, otherPlayer t3.Move) error {
	_, err := f.internalBoard.Place(otherPlayer)
	return err
}

func (f *FillInBotSession) NextMove(ctx context.Context) (out t3.Move, err error) {
	for {
		if f.nextPosition.Row >= 3 {
			return out, errors.New("no positions left")
		}

		occupied, err := f.internalBoard.Occupied(f.nextPosition)
		if err != nil {
			return out, err
		}
		if occupied == 0 {
			play := f.nextPosition
			f.nextPosition.Column++
			if f.nextPosition.Column >= 3 {
				f.nextPosition.Row++
				f.nextPosition.Column = 0
			}
			return play, nil
		} else {
			f.nextPosition.Column++
			if f.nextPosition.Column >= 3 {
				f.nextPosition.Row++
				f.nextPosition.Column = 0
			}
		}
	}
}

func (f *FillInBotSession) Done(ctx context.Context) error {
	return nil
}

func NewFillInBot() *FillInBotSession {
	return &FillInBotSession{
		internalBoard: t3.NewBoard(),
		nextPosition:  t3.Move{Row: 0, Column: 0},
	}
}
