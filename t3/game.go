package t3

import (
	"context"
	"errors"
)

type GameState int

const (
	GameStatePreStart = iota
	GameStatePlayer1Turn
	GameStatePlayer2Turn
	GateStateConcluded
)

// T3Game is a variation of Tic-Tac-Toe.
type T3Game struct {
	currentState GameState
	board        *T3Board
	p1           *T3Player
	p2           *T3Player
	// 0 = cats game, 0 > is player ID
	winner int
}

func NewT3Game(player1 *T3Player, player2 *T3Player) *T3Game {
	return &T3Game{
		currentState: GameStatePreStart,
		board:        NewT3Board(),
		p1:           player1,
		p2:           player2,
	}
}

func (t *T3Game) Step(ctx context.Context) (bool, error) {
	switch t.currentState {
	case GameStatePreStart:
		t.currentState = GameStatePlayer1Turn
	case GameStatePlayer1Turn:
		if err := t.doPlayerTurn(ctx, 1, t.p1); err != nil {
			return true, err
		}
		if t.board.completed(1) {
			t.currentState = GateStateConcluded
			t.winner = 1
			return false, nil
		} else {
			t.currentState = GameStatePlayer2Turn
		}
	case GameStatePlayer2Turn:
		if err := t.doPlayerTurn(ctx, 2, t.p2); err != nil {
			return true, err
		}
		if t.board.completed(2) {
			t.currentState = GateStateConcluded
			t.winner = 2
			return false, nil
		} else {
			t.currentState = GameStatePlayer1Turn
		}
	case GateStateConcluded:
		return false, nil
	default:
		return false, UnhandledGameState
	}
	return true, nil
}

func (t *T3Game) doPlayerTurn(ctx context.Context, side int, input *T3Player) error {
	move, err := input.NextPlay(ctx)
	if err != nil {
		return err
	}
	move.Player = side
	occupied, err := t.board.place(move)
	if err != nil {
		return err
	}
	if occupied { //todo: figure out how to handle bad plays...giving up turn sufficient?

	}
	return nil
}

var UnhandledGameState = errors.New("unhandled game state")
