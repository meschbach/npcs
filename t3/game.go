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

// Game is a variation of Tic-Tac-Toe.
type Game struct {
	currentState GameState
	board        *Board
	p1           Player
	p2           Player
	// 0 = cats game, 0 > is player ID
	winner int
}

func NewGame(player1 Player, player2 Player) *Game {
	return &Game{
		currentState: GameStatePreStart,
		board:        NewBoard(),
		p1:           player1,
		p2:           player2,
	}
}

func (t *Game) Step(ctx context.Context) error {
	switch t.currentState {
	case GameStatePreStart:
		t.currentState = GameStatePlayer1Turn
		return nil
	case GameStatePlayer1Turn:
		if err := t.doPlayerTurn(ctx, 1, t.p1); err != nil {
			return err
		}
		if t.board.completed(1) {
			t.currentState = GateStateConcluded
			t.winner = 1
		} else {
			t.currentState = GameStatePlayer2Turn
		}
	case GameStatePlayer2Turn:
		if err := t.doPlayerTurn(ctx, 2, t.p2); err != nil {
			return err
		}
		if t.board.completed(2) {
			t.currentState = GateStateConcluded
			t.winner = 2
		} else {
			t.currentState = GameStatePlayer1Turn
		}
	case GateStateConcluded:
	default:
		return UnhandledGameState
	}
	return nil
}

func (t *Game) Concluded() bool {
	return t.currentState == GateStateConcluded
}

func (t *Game) doPlayerTurn(ctx context.Context, side int, input Player) error {
	move, err := input.NextPlay(ctx)
	if err != nil {
		return &PlayerError{
			WhichPlayer: side,
			Performing:  "requesting next play",
			Underlying:  err,
		}
	}
	move.Player = side
	occupied, err := t.board.Place(move)
	if err != nil {
		return &PlayerError{
			WhichPlayer: side,
			Performing:  "placing",
			Underlying:  err,
		}
	}
	if occupied { //todo: figure out how to handle bad plays...giving up turn sufficient?

	}
	p1Err := t.p1.PushHistory(ctx, move)
	p2Err := t.p2.PushHistory(ctx, move)
	return errors.Join(p1Err, p2Err)
}

func (t *Game) Result() (concluded bool, winner int) {
	switch t.currentState {
	case GateStateConcluded:
		return true, t.winner
	default:
		return false, 0
	}
}

var UnhandledGameState = errors.New("unhandled game state")
