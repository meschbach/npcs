package integtest

import (
	"context"
	"github.com/meschbach/npcs/competition"
	"github.com/meschbach/npcs/competition/simple"
	"github.com/meschbach/npcs/junk/inProc"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSimpleOneOffGame(t *testing.T) {
	sysContext, sysDone := context.WithTimeout(context.Background(), 10*time.Second)
	defer sysDone()

	matcherAddress := "competition.npcs:12345"
	net := inProc.NewGRPCNetwork(t)
	competitionSystem := competition.NewCompetitionSystem(nil, matcherAddress, net, nil)
	go func() {
		require.NoError(t, competitionSystem.Serve(sysContext))
	}()
	competitionSystem.WaitForReady()

	instance := simple.NewRunOnceInstance(
		simple.WithInstanceNetwork(net),
		simple.WithInstanceAddress("game-1.simple.npcs:12345"),
		simple.WithInstanceMatcherAddress(matcherAddress),
	)
	require.NoError(t, instance.Run(sysContext))

	player1 := simple.NewRunOnce(simple.WithPlayerNetwork(net), simple.WithPlayerMatcherAddress(matcherAddress))
	go func() { require.NoError(t, player1.Run(sysContext)) }()
	player2 := simple.NewRunOnce(simple.WithPlayerNetwork(net), simple.WithPlayerMatcherAddress(matcherAddress))
	go func() { require.NoError(t, player2.Run(sysContext)) }()
}
