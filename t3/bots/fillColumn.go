package bots

import (
	"context"
	"errors"
	"github.com/meschbach/npcs/t3"
)

type FillColumn struct {
	nextPosition  t3.Move
	internalBoard *t3.Board
}

func NewFillColumn(start t3.Move) *FillColumn {
	return &FillColumn{
		nextPosition:  start,
		internalBoard: t3.NewBoard(),
	}
}

func (f *FillColumn) MoveMade(ctx context.Context, otherPlayer t3.Move) error {
	_, err := f.internalBoard.Place(otherPlayer)
	return err
}

func (f *FillColumn) NextMove(ctx context.Context) (out t3.Move, err error) {
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
			f.nextPosition.Row++
			return play, nil
		} else {
			return t3.Move{}, errors.New("no plays left")
		}
	}
}

func (f *FillColumn) Done(ctx context.Context) error {
	return nil
}
