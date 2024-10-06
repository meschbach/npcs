package npcs

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestT3Game(t *testing.T) {
	player1In := make(chan T3Move, 9)
	player1In <- T3Move{Row: 0, Column: 0}
	player1In <- T3Move{Row: 1, Column: 0}
	player1In <- T3Move{Row: 2, Column: 0}

	player2In := make(chan T3Move, 9)
	player2In <- T3Move{Row: 0, Column: 1}
	player2In <- T3Move{Row: 1, Column: 1}

	game := NewT3Game(player1In, player2In)

	t.Run("When asked complete 4 turns", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		t.Cleanup(cancel)

		for moves := 4; moves >= 0; moves-- {
			stillPlaying, err := game.Step(ctx)
			require.NoError(t, err)
			assert.True(t, stillPlaying, "still playing")
		}

		t.Run("Then the last turn marks the game as complete", func(t *testing.T) {
			stillPlaying, err := game.Step(ctx)
			require.NoError(t, err)
			assert.False(t, stillPlaying, "game should have completed")
			assert.Equal(t, 1, game.winner)
		})
	})
}
