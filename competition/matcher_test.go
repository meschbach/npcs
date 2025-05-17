package competition

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestSimpleGameLifecycle(t *testing.T) {
	ctx, done := context.WithCancel(context.Background())
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
	for i := 0; i < 2; i = i + 1 {
		allJoined.Add(1)
		playerID := fmt.Sprintf("%s-%d", baseUserName, i)
		go func() {
			_, err = core.findMatchInstance(ctx, exampleGame, playerID)
			allJoined.Done()
			require.NoError(t, err, "matcher should not have failed on iteration %d", i)
		}()
	}
	allJoined.Wait()

	require.NoError(t, core.gameCompleted(ctx, exampleGame, exampleInstanceID, baseUserName+"-0"))

	gameHistory := core.findAllGamesForPlayer(ctx, baseUserName+"-0")
	assert.Equal(t, 1, len(gameHistory))
}
