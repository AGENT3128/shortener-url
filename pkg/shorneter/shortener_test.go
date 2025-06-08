package shorneter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AGENT3128/shortener-url/pkg/shorneter"
)

func TestGenerateShortID(t *testing.T) {
	t.Run("length check", func(t *testing.T) {
		shortID := shorneter.GenerateShortID()
		assert.Len(t, shortID, shorneter.LENGTH, "Expected length of %d, got %d", shorneter.LENGTH, len(shortID))
	})

	t.Run("uniqueness check", func(t *testing.T) {
		iterations := 1000
		ids := make(map[string]bool)

		for range iterations {
			id := shorneter.GenerateShortID()
			assert.False(t, ids[id], "Duplicate ID found: %s", id)
			ids[id] = true
		}
	})
}

func TestGenerateShortIDOptimized(t *testing.T) {
	t.Run("length check", func(t *testing.T) {
		shortID, err := shorneter.GenerateShortIDOptimized()
		require.NoError(t, err)
		assert.Len(t, shortID, shorneter.LENGTH, "Expected length of %d, got %d", shorneter.LENGTH, len(shortID))
	})

	t.Run("uniqueness check", func(t *testing.T) {
		iterations := 1000
		ids := make(map[string]bool)

		for range iterations {
			id := shorneter.GenerateShortID()
			assert.False(t, ids[id], "Duplicate ID found: %s", id)
			ids[id] = true
		}
	})
}

func BenchmarkGenerateShortID(b *testing.B) {
	b.Run("original", func(b *testing.B) {
		for range b.N {
			shorneter.GenerateShortID()
		}
	})
	b.Run("optimized", func(b *testing.B) {
		for range b.N {
			shorneter.GenerateShortIDOptimized()
		}
	})
}
