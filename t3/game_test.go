package t3

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGame(t *testing.T) {
	player1In := make(chan Move, 9)
	player1In <- Move{Row: 0, Column: 0}
	player1In <- Move{Row: 1, Column: 0}
	player1In <- Move{Row: 2, Column: 0}

	player2In := make(chan Move, 9)
	player2In <- Move{Row: 0, Column: 1}
	player2In <- Move{Row: 1, Column: 1}

	game := NewGame(NewPlayer(player1In), NewPlayer(player2In))

	t.Run("When asked complete 4 turns", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		t.Cleanup(cancel)

		for moves := 4; moves >= 0; moves-- {
			require.NoError(t, game.Step(ctx))
			assert.False(t, game.Concluded(), "still playing")
		}

		t.Run("Then the last turn marks the game as complete", func(t *testing.T) {
			require.NoError(t, game.Step(ctx))
			assert.True(t, game.Concluded(), "game should have completed")
			assert.Equal(t, 1, game.winner)
		})
	})
}
