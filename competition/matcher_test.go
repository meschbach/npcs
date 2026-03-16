package competition

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleGameLifecycle(t *testing.T) {
	t.Parallel()
	ctx, done := context.WithCancel(t.Context())
	t.Cleanup(done)

	exampleGame := "test-game"
	exampleInstanceID := "ad-14-a-f"
	core := newMatcher()
	require.NoError(t, core.ensureGame(ctx, exampleGame))
	id, err := core.registerInstance(ctx, exampleGame, exampleInstanceID, "example://authority/resource")
	require.NoError(t, err)
	assert.Equal(t, "ad-14-a-f", id)

	baseUserName := "user-"
	var allJoined sync.WaitGroup
	for i := 0; i < 2; i++ {
		allJoined.Add(1)
		playerID := fmt.Sprintf("%s-%d", baseUserName, i)
		go func() {
			_, err := core.findMatchInstance(ctx, exampleGame, playerID)
			allJoined.Done()
			assert.NoError(t, err, "matcher should not have failed on iteration %d", i)
		}()
	}
	allJoined.Wait()

	require.NoError(t, core.gameCompleted(ctx, exampleGame, exampleInstanceID, baseUserName+"-0"))

	gameHistory := core.findAllGamesForPlayer(ctx, baseUserName+"-0")
	assert.Len(t, gameHistory, 1)
}
