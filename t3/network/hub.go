package network

import (
	"context"
	"errors"
	"github.com/meschbach/npcs/t3"
	"sync"
)

type GameFactory = func(ctx context.Context) (Session, error)

type Session interface {
	MoveMade(ctx context.Context, otherPlayer t3.Move) error
	NextMove(ctx context.Context) (t3.Move, error)
	Done(ctx context.Context) error
}

type Hub struct {
	UnimplementedT3Server
	newGame GameFactory

	lock   sync.Mutex
	lastID int64
	games  map[int64]Session
}

func NewHub(newGame GameFactory) *Hub {
	return &Hub{
		newGame: newGame,
		games:   make(map[int64]Session),
	}
}

func (h *Hub) StartGame(ctx context.Context, in *StartGameIn) (*StartGameOut, error) {
	s, err := h.newGame(ctx)
	if err != nil {
		return nil, err
	}

	h.lock.Lock()
	defer h.lock.Unlock()
	id := h.lastID
	h.lastID++
	h.games[id] = s

	return &StartGameOut{
		GameID: id,
	}, nil
}

func (h *Hub) findGame(gameID int64) (Session, error) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if game, has := h.games[gameID]; !has {
		return nil, errors.New("game not found")
	} else {
		return game, nil
	}
}

func (h *Hub) MoveMade(ctx context.Context, in *MoveMadeIn) (*MoveMadeOut, error) {
	game, err := h.findGame(in.GameID)
	if err != nil {
		return nil, err
	}
	err = game.MoveMade(ctx, t3.Move{
		Player: int(in.Player),
		Row:    int(in.Row),
		Column: int(in.Column),
	})
	return &MoveMadeOut{}, err
}

func (h *Hub) NextMove(ctx context.Context, in *NextMoveIn) (*NextMoveOut, error) {
	game, err := h.findGame(in.GameID)
	if err != nil {
		return nil, err
	}
	move, err := game.NextMove(ctx)
	if err != nil {
		return nil, err
	}
	return &NextMoveOut{
		Row:    int64(move.Row),
		Column: int64(move.Column),
	}, nil
}

func (h *Hub) Concluded(ctx context.Context, in *ConclusionIn) (*ConclusionOut, error) {
	h.lock.Lock()
	defer h.lock.Unlock()
	var err error
	if game, has := h.games[in.GameID]; !has {
		return nil, errors.New("game not found")
	} else {
		err = game.Done(ctx)
		delete(h.games, in.GameID)
	}
	return &ConclusionOut{}, err
}
