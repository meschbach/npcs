package competition

import "context"

type t3Game struct {
}

func newT3Game() *t3Game {
	return &t3Game{}
}

func (t *t3Game) Serve(ctx context.Context) error {
	return nil
}
