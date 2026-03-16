package integtest

import (
	"context"
	"testing"
	"time"

	"github.com/meschbach/npcs/competition"
	"github.com/meschbach/npcs/competition/simple"
	"github.com/meschbach/npcs/competition/wire"
	"github.com/meschbach/npcs/junk/inproc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleOneOffGame(t *testing.T) {
	t.Parallel()
	sysContext, sysDone := context.WithTimeout(t.Context(), 5*time.Second)
	defer sysDone()

	matcherAddress := "competition.npcs:12345"
	net := inproc.NewGRPCNetwork(t)
	competitionSystem := competition.NewCompetitionSystem(nil, matcherAddress, net, nil)
	go func() {
		assert.NoError(t, competitionSystem.Serve(sysContext))
	}()
	competitionSystem.WaitForReady()

	instance := simple.NewRunOnceInstance(
		simple.WithInstanceNetwork(net),
		simple.WithInstanceAddress("game-1.simple.npcs:12345"),
		simple.WithInstanceMatcherAddress(matcherAddress),
	)
	require.NoError(t, instance.Run(sysContext))
	instance.WaitForStartup()

	player1 := simple.NewRunOnce(simple.WithPlayerNetwork(net), simple.WithPlayerMatcherAddress(matcherAddress))
	go func() { assert.NoError(t, player1.Run(sysContext)) }()
	player2 := simple.NewRunOnce(simple.WithPlayerNetwork(net), simple.WithPlayerMatcherAddress(matcherAddress))
	go func() { assert.NoError(t, player2.Run(sysContext)) }()

	//todo: figure out how to sync with matcher for game completion
	require.NoError(t, instance.WaitForCompletion(sysContext))

	matcherGRPC, err := net.Client(sysContext, matcherAddress)
	require.NoError(t, err)
	matcherClient := wire.NewCompetitionV1Client(matcherGRPC)
	history, err := matcherClient.GetHistory(sysContext, &wire.RecordIn{ForPlayer: "test-1234"})
	require.NoError(t, err)
	if assert.Len(t, history.Games, 1) {
		assert.Equal(t, "github.com/meschbach/npc/competition/simple/v0", history.Games[0].Game)
	}
}
