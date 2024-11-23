package integtest

import (
	"context"
	"github.com/meschbach/npcs/competition/wire"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSimpleOneOffGame(t *testing.T) {
	WithCompetitionSystem(t, func(ctx context.Context, t *testing.T, h *Harness) {
		player1Token, _ := h.Auth.NewUser(ctx, t, "player 1")

		player1 := h.NewClient(player1Token)
		_, err := player1.RegisterPersistentPlayer(ctx, &wire.RegisterPlayerIn{
			OrchestrationURL: "player-1.npcs:1234",
			Name:             "persistent",
		})
		require.NoError(t, err)
		result, err := player1.QuickMatch(ctx, &wire.QuickMatchIn{})
		require.NoError(t, err)

		require.NotNil(t, result)
		assert.NotEmpty(t, result.MatchURL)

		//dial out
		//play the game
	})
}
