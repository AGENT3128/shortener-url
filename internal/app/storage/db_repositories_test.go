package storage

import (
	"testing"

	"github.com/AGENT3128/shortener-url/internal/app/storage/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestMockDBStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)

	tests := []struct {
		name        string
		shortID     string
		originalURL string
	}{
		{
			name:        "add new original URL1",
			shortID:     "abc123",
			originalURL: "https://ya.ru",
		},
		{
			name:        "add new original URL2",
			shortID:     "abc456",
			originalURL: "https://yandex.ru",
		},
		{
			name:        "add duplicate original URL2",
			shortID:     "abc456",
			originalURL: "https://yandex.ru",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.EXPECT().Add(tt.shortID, tt.originalURL)

			mockRepo.EXPECT().
				GetByShortID(tt.shortID).
				Return(tt.originalURL, true)

			mockRepo.EXPECT().
				GetByOriginalURL(tt.originalURL).
				Return(tt.shortID, true)

			mockRepo.Add(tt.shortID, tt.originalURL)

			originalURL, ok := mockRepo.GetByShortID(tt.shortID)
			assert.True(t, ok)
			assert.Equal(t, tt.originalURL, originalURL)

			shortID, ok := mockRepo.GetByOriginalURL(tt.originalURL)
			assert.True(t, ok)
			assert.Equal(t, tt.shortID, shortID)
		})
	}
}
