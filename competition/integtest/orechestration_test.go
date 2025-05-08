package integtest

import (
	"context"
	"github.com/meschbach/npcs/competition/wire"
	"github.com/meschbach/npcs/t3/bots"
	t3net "github.com/meschbach/npcs/t3/network"
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
		result, err := player1.QuickMatch(ctx, &wire.QuickMatchIn{})
		require.NoError(t, err)

		require.NotNil(t, result)
		assert.NotEmpty(t, result.MatchURL)

		// and construct a push client
		fillIn := bots.NewFillInBot()
		client := t3net.NewPushClient("in-proc://"+result.MatchURL, result.UUID, player1Token, fillIn, h.NewGRPCClientOptions("competition.npcs", result.MatchURL)...)
		// then we should be able to play through the game.
		clientError := client.Serve(ctx)
		require.NoError(t, clientError)
	})
}
