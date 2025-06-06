package shorneter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateShortID(t *testing.T) {
	t.Run("length check", func(t *testing.T) {
		shortID := GenerateShortID()
		assert.Len(t, shortID, LENGTH, "Expected length of %d, got %d", LENGTH, len(shortID))
	})

	t.Run("uniqueness check", func(t *testing.T) {
		iterations := 1000
		ids := make(map[string]bool)

		for range iterations {
			id := GenerateShortID()
			assert.False(t, ids[id], "Duplicate ID found: %s", id)
			ids[id] = true
		}
	})
}
