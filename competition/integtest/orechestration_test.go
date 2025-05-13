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
		//Given an existing player
		player1Token, _ := h.Auth.NewUser(ctx, t, "player 1")
		player1 := h.NewClient(player1Token)

		// with a registered persistent player
		_, err := player1.RegisterPersistentPlayer(ctx, &wire.RegisterPlayerIn{
			OrchestrationURL: "player-1.npcs:1234",
			Name:             "persistent",
		})
		require.NoError(t, err)

		/// When we attempt to schedule a quick match
		quickMatchName := "quick-match-player"
		result, err := player1.QuickMatch(ctx, &wire.QuickMatchIn{
			PlayerName: quickMatchName,
		})
		require.NoError(t, err)

		require.NotNil(t, result)
		assert.NotEmpty(t, result.MatchURL)

		// and constructs a new simple test service client
		simpleGameAddress := h.internet.Connect(ctx, "in-proc://"+result.MatchURL)
		simpleGameClient := wire.NewSimpleTestGameServiceClient(simpleGameAddress)
		_, err = simpleGameClient.Connected(ctx, &wire.SimpleTestGameIn{GameID: result.UUID})
		require.NoError(t, err)

		// then the quick match name should have been recorded as the winner
		gameResult, err := player1.GameResult(ctx, &wire.GameResultIn{GameID: result.UUID})
		require.NoError(t, err)
		assert.Equal(t, gameResult.WinningPlayerName, quickMatchName)
	})
}
