package network

import (
	"context"
	"github.com/meschbach/npcs/t3"
)

type LocalSession struct {
	s Session
}

func (l LocalSession) NextPlay(ctx context.Context) (t3.Move, error) {
	return l.s.NextMove(ctx)
}

func (l LocalSession) PushHistory(ctx context.Context, move t3.Move) error {
	return l.s.MoveMade(ctx, move)
}

func WithLocalSession(s Session) *LocalSession {
	return &LocalSession{s}
}
