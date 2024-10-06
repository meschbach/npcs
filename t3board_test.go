package npcs

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestT3Board(t *testing.T) {
	t.Run("Empty board should not be completed", func(t *testing.T) {
		b := NewT3Board()
		assert.False(t, b.completed(1))
		assert.False(t, b.completed(2))
	})

	t.Run("Given player 1 has the first row", func(t *testing.T) {
		b := NewT3Board()
		for index := 0; index < 3; index++ {
			_, err := b.place(T3Move{Player: 1, Row: 0, Column: index})
			require.NoError(t, err)
		}
		assert.True(t, b.completed(1))
	})

	t.Run("Given player 2 has the 2nd column", func(t *testing.T) {
		b := NewT3Board()
		for index := 0; index < 3; index++ {
			_, err := b.place(T3Move{Player: 2, Row: index, Column: 1})
			require.NoError(t, err)
		}
		assert.True(t, b.completed(2))
	})

	t.Run("Given player 3 has a diagonal top lef to bottom right", func(t *testing.T) {
		b := NewT3Board()
		for index := 0; index < 3; index++ {
			blocked, err := b.place(T3Move{Player: 3, Row: index, Column: index})
			assert.False(t, blocked)
			require.NoError(t, err)
		}
		assert.True(t, b.completed(3))
	})
	t.Run("Given player 4 has a diagonal bottom left to top right", func(t *testing.T) {
		b := NewT3Board()
		for index := 0; index < 3; index++ {
			blocked, err := b.place(T3Move{Player: 4, Row: 2 - index, Column: index})
			assert.False(t, blocked)
			require.NoError(t, err)
		}
		assert.True(t, b.completed(4))
	})
}
