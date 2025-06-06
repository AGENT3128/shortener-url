package usecase

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/entity"
	"github.com/AGENT3128/shortener-url/internal/worker"
	"github.com/AGENT3128/shortener-url/pkg/shorneter"
)

type options struct {
	repository URLRepository
	logger     *zap.Logger
	worker     *worker.DeleteWorker
}

// Option is the option for the URLUsecase.
type Option func(options *options) error

// URLUsecase is the usecase for the URL.
type URLUsecase struct {
	repository URLRepository
	logger     *zap.Logger
	worker     *worker.DeleteWorker
}

// NewURLUsecase creates a new URLUsecase.
func NewURLUsecase(opts ...Option) (*URLUsecase, error) {
	options := &options{}
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, err
		}
	}
	if options.repository == nil {
		return nil, errors.New("repository is required")
	}
	if options.logger == nil {
		return nil, errors.New("logger is required")
	}
	return &URLUsecase{
		repository: options.repository,
		logger:     options.logger,
		worker:     options.worker,
	}, nil
}

// WithDeleteWorker is the option for the URLUsecase to set the delete worker.
func WithDeleteWorker(worker *worker.DeleteWorker) Option {
	return func(options *options) error {
		options.worker = worker
		return nil
	}
}

// WithURLUsecaseRepository is the option for the URLUsecase to set the repository.
func WithURLUsecaseRepository(repository URLRepository) Option {
	return func(options *options) error {
		options.repository = repository
		return nil
	}
}

// WithURLUsecaseLogger is the option for the URLUsecase to set the logger.
func WithURLUsecaseLogger(logger *zap.Logger) Option {
	return func(options *options) error {
		options.logger = logger.With(zap.String("usecase", "URLUsecase"))
		return nil
	}
}

// Shutdown shuts down the URLUsecase.
func (uc *URLUsecase) Shutdown() {
	if uc.worker != nil {
		uc.worker.Shutdown()
	}
}

// Add adds a URL.
func (uc *URLUsecase) Add(ctx context.Context, userID string, originalURL string) (string, error) {
	shortURL, err := shorneter.GenerateShortIDOptimized()
	if err != nil {
		return "", err
	}
	shortURL, err = uc.repository.Add(ctx, userID, shortURL, originalURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			// Try to get existing short URL for this original URL
			existingShortURL, getErr := uc.GetByOriginalURL(ctx, originalURL)
			if getErr != nil {
				return "", err
			}
			return existingShortURL, entity.ErrURLExists
		}
		return "", err
	}

	return shortURL, nil
}

// GetByOriginalURL gets the short URL by the original URL.
func (uc *URLUsecase) GetByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	shortURL, err := uc.repository.GetByOriginalURL(ctx, originalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", entity.ErrURLNotFound
		}
		return "", err
	}
	return shortURL, nil
}

// GetByShortURL gets the original URL by the short URL.
func (uc *URLUsecase) GetByShortURL(ctx context.Context, shortURL string) (string, error) {
	uc.logger.Info("searching for short URL", zap.String("short_url", shortURL))
	originalURL, err := uc.repository.GetByShortURL(ctx, shortURL)
	if err != nil {
		return "", err
	}
	uc.logger.Info("short URL found", zap.String("short_url", shortURL), zap.String("original_url", originalURL))
	return originalURL, nil
}

// Ping pings the repository.
func (uc *URLUsecase) Ping(ctx context.Context) error {
	return uc.repository.Ping(ctx)
}

// AddBatch adds a batch of URLs.
func (uc *URLUsecase) AddBatch(ctx context.Context, userID string, urls []entity.URL) ([]entity.URL, error) {
	uc.logger.Info("adding batch of URLs", zap.Any("urls", urls))
	uniqueURLs := make([]entity.URL, 0, len(urls))
	result := make([]entity.URL, 0, len(urls))
	for _, url := range urls {
		uc.logger.Info("processing url", zap.String("url", url.OriginalURL))
		// if OriginalURL exist in db
		existingURL, err := uc.GetByOriginalURL(ctx, url.OriginalURL)
		if err != nil {
			if errors.Is(err, entity.ErrURLNotFound) {
				shortURL, errGenerate := shorneter.GenerateShortIDOptimized()
				if errGenerate != nil {
					return nil, errGenerate
				}
				uniqueURLs = append(uniqueURLs, entity.URL{
					OriginalURL: url.OriginalURL,
					ShortURL:    shortURL,
				})
				result = append(result, entity.URL{
					OriginalURL: url.OriginalURL,
					ShortURL:    shortURL,
				})
				continue
			}
			return nil, err
		}
		result = append(result, entity.URL{
			OriginalURL: url.OriginalURL,
			ShortURL:    existingURL,
		})
	}

	uc.logger.Info("response", zap.Any("response", result))
	return result, uc.repository.AddBatch(ctx, userID, uniqueURLs)
}

// GetUserURLs gets user URLs.
func (uc *URLUsecase) GetUserURLs(ctx context.Context, userID string) ([]entity.URL, error) {
	return uc.repository.GetUserURLs(ctx, userID)
}

// DeleteUserURLs deletes user URLs.
func (uc *URLUsecase) DeleteUserURLs(_ context.Context, userID string, shortURLs []string) error {
	uc.logger.Info("deleting user URLs", zap.String("userID", userID), zap.Any("shortURLs", shortURLs))
	if uc.worker != nil {
		uc.worker.EnqueueDelete(worker.DeleteRequest{
			UserID:    userID,
			ShortURLs: shortURLs,
		})
	}
	return nil
}
