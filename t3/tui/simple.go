package tui

import (
	"context"
	"fmt"
	"github.com/meschbach/npcs/t3"
)

type simple struct {
	player int64
}

func (s simple) NextPlay(ctx context.Context) (out t3.Move, err error) {
	fmt.Printf("Place at (column,row)?\t")
	_, err = fmt.Scanf("%d %d\n", &out.Column, &out.Row)
	return out, err
}

func (s simple) PushHistory(ctx context.Context, move t3.Move) error {
	fmt.Printf("Player %d moved to %d,%d\n", move.Player, move.Column, move.Row)
	return nil
}
