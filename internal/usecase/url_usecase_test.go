package usecase_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/entity"
	"github.com/AGENT3128/shortener-url/internal/usecase"
	"github.com/AGENT3128/shortener-url/internal/usecase/mocks"
	"github.com/AGENT3128/shortener-url/internal/worker"
)

func TestURLUsecase_Add(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	urlRepositoryMock := mocks.NewMockURLRepository(ctrl)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	usecase, err := usecase.NewURLUsecase(
		usecase.WithURLUsecaseRepository(urlRepositoryMock),
		usecase.WithURLUsecaseLogger(logger),
	)
	require.NoError(t, err)

	tests := []struct {
		errType error
		url     *entity.URL
		setup   func()
		name    string
		want    string
		wantErr bool
	}{
		{
			name: "success save url",
			url: &entity.URL{
				ShortURL:    "mEENY1b2",
				UserID:      "1",
				OriginalURL: "https://example.com",
			},
			setup: func() {
				urlRepositoryMock.EXPECT().
					Add(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("mEENY1b2", nil)
			},
			wantErr: false,
			want:    "mEENY1b2",
		},
		{
			name: "invalid save dublicate url",
			url: &entity.URL{
				ShortURL:    "mEENY1b2",
				UserID:      "1",
				OriginalURL: "https://example.com",
			},
			setup: func() {
				urlRepositoryMock.EXPECT().
					Add(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", entity.ErrURLExists)
			},
			wantErr: true,
			errType: entity.ErrURLExists,
			want:    "",
		},
		{
			name: "postgres unique violation with successful get existing url",
			url: &entity.URL{
				ShortURL:    "mEENY1b2",
				UserID:      "1",
				OriginalURL: "https://example.com",
			},
			setup: func() {
				pgErr := &pgconn.PgError{Code: pgerrcode.UniqueViolation}
				urlRepositoryMock.EXPECT().
					Add(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", pgErr)
				urlRepositoryMock.EXPECT().
					GetByOriginalURL(gomock.Any(), "https://example.com").
					Return("existingShort", nil)
			},
			wantErr: true,
			errType: entity.ErrURLExists,
			want:    "existingShort",
		},
		{
			name: "postgres unique violation with failed get existing url",
			url: &entity.URL{
				ShortURL:    "mEENY1b2",
				UserID:      "1",
				OriginalURL: "https://example.com",
			},
			setup: func() {
				pgErr := &pgconn.PgError{Code: pgerrcode.UniqueViolation}
				urlRepositoryMock.EXPECT().
					Add(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", pgErr)
				urlRepositoryMock.EXPECT().
					GetByOriginalURL(gomock.Any(), "https://example.com").
					Return("", errors.New("db error"))
			},
			wantErr: true,
			want:    "",
		},
		{
			name: "other repository error",
			url: &entity.URL{
				ShortURL:    "mEENY1b2",
				UserID:      "1",
				OriginalURL: "https://example.com",
			},
			setup: func() {
				urlRepositoryMock.EXPECT().
					Add(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", errors.New("repository error"))
			},
			wantErr: true,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			got, errAdd := usecase.Add(ctx, tt.url.UserID, tt.url.OriginalURL)

			if tt.wantErr {
				require.Error(t, errAdd)
				if tt.errType != nil {
					require.ErrorIs(t, errAdd, tt.errType)
				}
			} else {
				require.NoError(t, errAdd)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestURLUsecase_GetByOriginalURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	urlRepositoryMock := mocks.NewMockURLRepository(ctrl)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	usecase, err := usecase.NewURLUsecase(
		usecase.WithURLUsecaseRepository(urlRepositoryMock),
		usecase.WithURLUsecaseLogger(logger),
	)
	require.NoError(t, err)

	tests := []struct {
		errType     error
		setup       func()
		name        string
		originalURL string
		want        string
		wantErr     bool
	}{
		{
			name:        "success get url",
			originalURL: "https://example.com",
			setup: func() {
				urlRepositoryMock.EXPECT().
					GetByOriginalURL(gomock.Any(), "https://example.com").
					Return("shortURL", nil)
			},
			want:    "shortURL",
			wantErr: false,
		},
		{
			name:        "url not found",
			originalURL: "https://example.com",
			setup: func() {
				urlRepositoryMock.EXPECT().
					GetByOriginalURL(gomock.Any(), "https://example.com").
					Return("", sql.ErrNoRows)
			},
			want:    "",
			wantErr: true,
			errType: entity.ErrURLNotFound,
		},
		{
			name:        "repository error",
			originalURL: "https://example.com",
			setup: func() {
				urlRepositoryMock.EXPECT().
					GetByOriginalURL(gomock.Any(), "https://example.com").
					Return("", errors.New("repository error"))
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			got, errGet := usecase.GetByOriginalURL(ctx, tt.originalURL)
			if tt.wantErr {
				require.Error(t, errGet)
				if tt.errType != nil {
					require.ErrorIs(t, errGet, tt.errType)
				}
			} else {
				require.NoError(t, errGet)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestURLUsecase_GetByShortURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	urlRepositoryMock := mocks.NewMockURLRepository(ctrl)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	usecase, err := usecase.NewURLUsecase(
		usecase.WithURLUsecaseRepository(urlRepositoryMock),
		usecase.WithURLUsecaseLogger(logger),
	)
	require.NoError(t, err)

	tests := []struct {
		name     string
		shortURL string
		setup    func()
		want     string
		wantErr  bool
	}{
		{
			name:     "success get url",
			shortURL: "abc123",
			setup: func() {
				urlRepositoryMock.EXPECT().
					GetByShortURL(gomock.Any(), "abc123").
					Return("https://example.com", nil)
			},
			want:    "https://example.com",
			wantErr: false,
		},
		{
			name:     "url not found",
			shortURL: "abc123",
			setup: func() {
				urlRepositoryMock.EXPECT().
					GetByShortURL(gomock.Any(), "abc123").
					Return("", sql.ErrNoRows)
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			got, errGet := usecase.GetByShortURL(ctx, tt.shortURL)
			if tt.wantErr {
				require.Error(t, errGet)
			} else {
				require.NoError(t, errGet)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestURLUsecase_AddBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	urlRepositoryMock := mocks.NewMockURLRepository(ctrl)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	usecase, err := usecase.NewURLUsecase(
		usecase.WithURLUsecaseRepository(urlRepositoryMock),
		usecase.WithURLUsecaseLogger(logger),
	)
	require.NoError(t, err)

	tests := []struct {
		name    string
		userID  string
		urls    []entity.URL
		setup   func()
		want    []entity.URL
		wantErr bool
	}{
		{
			name:   "success add batch",
			userID: "user1",
			urls: []entity.URL{
				{OriginalURL: "https://example1.com"},
				{OriginalURL: "https://example2.com"},
			},
			setup: func() {
				urlRepositoryMock.EXPECT().
					GetByOriginalURL(gomock.Any(), "https://example1.com").
					Return("", entity.ErrURLNotFound)
				urlRepositoryMock.EXPECT().
					GetByOriginalURL(gomock.Any(), "https://example2.com").
					Return("", entity.ErrURLNotFound)
				urlRepositoryMock.EXPECT().
					AddBatch(gomock.Any(), "user1", gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "partial existing urls",
			userID: "user1",
			urls: []entity.URL{
				{OriginalURL: "https://example1.com"},
				{OriginalURL: "https://example2.com"},
			},
			setup: func() {
				urlRepositoryMock.EXPECT().
					GetByOriginalURL(gomock.Any(), "https://example1.com").
					Return("existing1", nil)
				urlRepositoryMock.EXPECT().
					GetByOriginalURL(gomock.Any(), "https://example2.com").
					Return("", entity.ErrURLNotFound)
				urlRepositoryMock.EXPECT().
					AddBatch(gomock.Any(), "user1", gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			got, errAdd := usecase.AddBatch(ctx, tt.userID, tt.urls)
			if tt.wantErr {
				require.Error(t, errAdd)
			} else {
				require.NoError(t, errAdd)
				require.Len(t, got, len(tt.urls))
			}
		})
	}
}

func TestURLUsecase_GetUserURLs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	urlRepositoryMock := mocks.NewMockURLRepository(ctrl)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	usecase, err := usecase.NewURLUsecase(
		usecase.WithURLUsecaseRepository(urlRepositoryMock),
		usecase.WithURLUsecaseLogger(logger),
	)
	require.NoError(t, err)

	tests := []struct {
		name    string
		userID  string
		setup   func()
		want    []entity.URL
		wantErr bool
	}{
		{
			name:   "success get user urls",
			userID: "user1",
			setup: func() {
				urlRepositoryMock.EXPECT().
					GetUserURLs(gomock.Any(), "user1").
					Return([]entity.URL{
						{ShortURL: "abc123", OriginalURL: "https://example1.com"},
						{ShortURL: "def456", OriginalURL: "https://example2.com"},
					}, nil)
			},
			want: []entity.URL{
				{ShortURL: "abc123", OriginalURL: "https://example1.com"},
				{ShortURL: "def456", OriginalURL: "https://example2.com"},
			},
			wantErr: false,
		},
		{
			name:   "repository error",
			userID: "user1",
			setup: func() {
				urlRepositoryMock.EXPECT().
					GetUserURLs(gomock.Any(), "user1").
					Return(nil, errors.New("repository error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			got, errGet := usecase.GetUserURLs(ctx, tt.userID)
			if tt.wantErr {
				require.Error(t, errGet)
			} else {
				require.NoError(t, errGet)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestURLUsecase_DeleteUserURLs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	urlRepositoryMock := mocks.NewMockURLRepository(ctrl)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	deleteWorker := worker.NewDeleteWorker(urlRepositoryMock, logger)

	usecase, err := usecase.NewURLUsecase(
		usecase.WithURLUsecaseRepository(urlRepositoryMock),
		usecase.WithURLUsecaseLogger(logger),
		usecase.WithDeleteWorker(deleteWorker),
	)
	require.NoError(t, err)

	tests := []struct {
		setup     func()
		name      string
		userID    string
		shortURLs []string
		wantErr   bool
	}{
		{
			name:      "success delete urls",
			userID:    "user1",
			shortURLs: []string{"abc123", "def456"},
			setup: func() {
				urlRepositoryMock.EXPECT().
					MarkDeletedBatch(gomock.Any(), "user1", []string{"abc123", "def456"}).
					Return(nil)
				urlRepositoryMock.EXPECT().
					Close().
					Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			errDelete := usecase.DeleteUserURLs(ctx, tt.userID, tt.shortURLs)
			if tt.wantErr {
				require.Error(t, errDelete)
			} else {
				require.NoError(t, errDelete)
			}
			time.Sleep(100 * time.Millisecond)
		})
	}

	usecase.Shutdown()
}

func TestURLUsecase_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	urlRepositoryMock := mocks.NewMockURLRepository(ctrl)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	usecase, err := usecase.NewURLUsecase(
		usecase.WithURLUsecaseRepository(urlRepositoryMock),
		usecase.WithURLUsecaseLogger(logger),
	)
	require.NoError(t, err)

	tests := []struct {
		setup   func()
		name    string
		wantErr bool
	}{
		{
			name: "success ping",
			setup: func() {
				urlRepositoryMock.EXPECT().
					Ping(gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repository error",
			setup: func() {
				urlRepositoryMock.EXPECT().
					Ping(gomock.Any()).
					Return(errors.New("repository error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			pingErr := usecase.Ping(ctx)
			if tt.wantErr {
				require.Error(t, pingErr)
			} else {
				require.NoError(t, pingErr)
			}
			time.Sleep(100 * time.Millisecond)
		})
	}
}
