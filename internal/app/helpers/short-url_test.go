package helpers

import "testing"

func TestGenerateShortID(t *testing.T) {
	t.Run("length check", func(t *testing.T) {
		shortID := GenerateShortID()
		if len(shortID) != LENGTH {
			t.Errorf("Expected length of %d, got %d", LENGTH, len(shortID))
		}
	})

	t.Run("uniqueness check", func(t *testing.T) {
		iterations := 1000
		ids := make(map[string]bool)

		for i := 0; i < iterations; i++ {
			id := GenerateShortID()
			if ids[id] {
				t.Errorf("Duplicate ID found: %s", id)
			}
			ids[id] = true
		}
	})
}
