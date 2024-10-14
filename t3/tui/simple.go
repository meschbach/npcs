package tui

import (
	"context"
	"errors"
	"fmt"
	"github.com/meschbach/npcs/t3"
)

type simple struct {
	player int64
	board  *t3.Board
}

func (s *simple) NextPlay(ctx context.Context) (out t3.Move, err error) {
	renderedBoard, err := renderBoardAsString(s.board)
	if err != nil {
		return t3.Move{}, err
	}
	fmt.Println(renderedBoard)
	fmt.Printf("Place at (column,row)?\t")
	_, err = fmt.Scanf("%d %d\n", &out.Column, &out.Row)
	return out, err
}

func (s *simple) PushHistory(ctx context.Context, move t3.Move) error {
	fmt.Printf("Player %d moved to %d,%d\n", move.Player, move.Column, move.Row)
	occupied, err := s.board.Place(move)
	if err != nil {
		return err
	}
	if occupied {
		return errors.New("consistency error: already occupied")
	}
	return nil
}

func (s *simple) concludedGame(winner int) {
	board, _ := renderBoardAsString(s.board)
	fmt.Printf("%s\nWinning player: %d\n", board, winner)
}
