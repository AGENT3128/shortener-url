package worker

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	defaultFlushInterval = 500 * time.Millisecond
	defaultBatchSize     = 50
	defaultChannelSize   = 100
)

// URLDeleter describes the behavior for marking URLs as deleted in batch.
type URLDeleter interface {
	MarkDeletedBatch(ctx context.Context, userID string, shortURLs []string) error
}

// DeleteRequest represents a request to delete URLs.
type DeleteRequest struct {
	UserID    string
	ShortURLs []string
}

// DeleteWorker handles batch processing of URL deletion requests.
type DeleteWorker struct {
	repository     URLDeleter
	logger         *zap.Logger
	deleteRequests chan DeleteRequest
	batchSize      int
	flushInterval  time.Duration
	done           chan struct{}
	wg             sync.WaitGroup
}

// NewDeleteWorker creates a new worker for processing delete requests.
func NewDeleteWorker(repo URLDeleter, logger *zap.Logger, opts ...Option) *DeleteWorker {
	logger = logger.With(zap.String("component", "DeleteWorker"))

	w := &DeleteWorker{
		repository:     repo,
		logger:         logger,
		deleteRequests: make(chan DeleteRequest, defaultChannelSize),
		batchSize:      defaultBatchSize,
		flushInterval:  defaultFlushInterval,
		done:           make(chan struct{}),
	}

	// Apply options
	for _, opt := range opts {
		opt(w)
	}

	w.wg.Add(1)
	go w.processDeleteRequests()

	return w
}

// Option is a function that configures DeleteWorker.
type Option func(*DeleteWorker)

// WithBatchSize sets the batch size for the worker.
func WithBatchSize(size int) Option {
	return func(w *DeleteWorker) {
		w.batchSize = size
	}
}

// WithFlushInterval sets the flush interval for the worker.
func WithFlushInterval(interval time.Duration) Option {
	return func(w *DeleteWorker) {
		w.flushInterval = interval
	}
}

// EnqueueDelete adds a delete request to the processing queue.
func (w *DeleteWorker) EnqueueDelete(req DeleteRequest) bool {
	select {
	case w.deleteRequests <- req:
		return true
	default:
		// Process immediately if channel is full
		go func(userID string, ids []string) {
			w.logger.Warn("delete requests channel is full, processing immediately",
				zap.String("userID", userID),
				zap.Int("urlCount", len(ids)))
			w.processBatch(userID, ids)
		}(req.UserID, req.ShortURLs)
		return false
	}
}

// processDeleteRequests implements the fanIn pattern for processing delete requests.
func (w *DeleteWorker) processDeleteRequests() {
	defer w.wg.Done()

	batches := make(map[string][]string)
	lastFlush := time.Now()

	const tickerInterval = 5
	ticker := time.NewTicker(w.flushInterval / tickerInterval)
	defer ticker.Stop()

	for {
		select {
		case req, ok := <-w.deleteRequests:
			if !ok {
				w.flushAllBatches(batches)
				return
			}

			batches[req.UserID] = append(batches[req.UserID], req.ShortURLs...)

			for userID, shortURLs := range batches {
				if len(shortURLs) >= w.batchSize {
					w.processBatch(userID, shortURLs)
					delete(batches, userID)
					lastFlush = time.Now()
				}
			}

		case <-ticker.C:
			if time.Since(lastFlush) >= w.flushInterval && len(batches) > 0 {
				w.flushAllBatches(batches)
				batches = make(map[string][]string)
				lastFlush = time.Now()
			}

		case <-w.done:
			w.logger.Info("Stopping URL delete processor")
			w.flushAllBatches(batches)
			return
		}
	}
}

// processBatch processes one batch of delete requests.
func (w *DeleteWorker) processBatch(userID string, shortURLs []string) {
	const timeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := w.repository.MarkDeletedBatch(ctx, userID, shortURLs)
	if err != nil {
		w.logger.Error("failed to mark URLs as deleted",
			zap.String("userID", userID),
			zap.Int("count", len(shortURLs)),
			zap.Error(err))
	} else {
		w.logger.Info("successfully marked URLs as deleted",
			zap.String("userID", userID),
			zap.Int("count", len(shortURLs)))
	}
}

// flushAllBatches processes all remaining batches.
func (w *DeleteWorker) flushAllBatches(batches map[string][]string) {
	for userID, shortURLs := range batches {
		if len(shortURLs) > 0 {
			w.processBatch(userID, shortURLs)
		}
	}
}

// Shutdown correctly stops the worker.
func (w *DeleteWorker) Shutdown() {
	w.logger.Info("Shutting down DeleteWorker")
	close(w.done)
	w.wg.Wait()
	w.logger.Info("DeleteWorker shutdown complete")
}
