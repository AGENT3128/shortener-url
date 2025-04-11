package storage

import (
	"context"
	"testing"
	"time"

	"github.com/AGENT3128/shortener-url/internal/app/storage/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMockDBStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	userID := uuid.New().String()

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
			mockRepo.EXPECT().Add(ctx, userID, tt.shortID, tt.originalURL)

			mockRepo.EXPECT().
				GetByShortID(ctx, tt.shortID).
				Return(tt.originalURL, true)

			mockRepo.EXPECT().
				GetByOriginalURL(ctx, tt.originalURL).
				Return(tt.shortID, true)

			mockRepo.Add(ctx, userID, tt.shortID, tt.originalURL)

			originalURL, ok := mockRepo.GetByShortID(ctx, tt.shortID)
			assert.True(t, ok)
			assert.Equal(t, tt.originalURL, originalURL)

			shortID, ok := mockRepo.GetByOriginalURL(ctx, tt.originalURL)
			assert.True(t, ok)
			assert.Equal(t, tt.shortID, shortID)
		})
	}
}
