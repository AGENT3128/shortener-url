package handlers

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// urlDeleter describes the behavior for marking URLs as deleted in batch
type urlDeleter interface {
	MarkDeletedBatch(ctx context.Context, userID string, shortIDs []string) error
}

// UserURLsDeleteHandler handles user URLs deletion requests
type UserURLsDeleteHandler struct {
	repository     urlDeleter
	logger         *zap.Logger
	deleteRequests chan deleteRequest
	// batch size
	batchSize int
	// flush interval if batch is not full
	flushInterval time.Duration
	// channel for signal about shutdown
	done chan struct{}
	wg   sync.WaitGroup
}

// deleteRequest represents a request to delete URLs
type deleteRequest struct {
	userID   string
	shortIDs []string
}

// NewUserURLsDeleteHandler creates a new handler for deleting URLs
func NewUserURLsDeleteHandler(repo urlDeleter, logger *zap.Logger) *UserURLsDeleteHandler {
	logger = logger.With(zap.String("handler", "UserURLsDeleteHandler"))

	h := &UserURLsDeleteHandler{
		repository:     repo,
		logger:         logger,
		deleteRequests: make(chan deleteRequest, 100),
		batchSize:      50,
		flushInterval:  500 * time.Millisecond,
		done:           make(chan struct{}),
	}

	h.wg.Add(1)
	go h.processDeleteRequests()

	return h
}

// Pattern returns the URL pattern for the handler
func (h *UserURLsDeleteHandler) Pattern() string {
	return "/api/user/urls"
}

// Method returns the HTTP method for the handler
func (h *UserURLsDeleteHandler) Method() string {
	return http.MethodDelete
}

// Handler returns the gin.HandlerFunc for the handler
func (h *UserURLsDeleteHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var shortIDs []string
		if err := c.ShouldBindJSON(&shortIDs); err != nil {
			h.logger.Error("failed to bind request body", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
			return
		}

		if len(shortIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no URLs provided for deletion"})
			return
		}

		select {
		// successfully sent to channel
		case h.deleteRequests <- deleteRequest{
			userID:   userID.(string),
			shortIDs: shortIDs,
		}:
		default:
			// channel is full, but we still need to process the request
			// try to process this batch directly rather than dropping it
			go func(userID string, ids []string) {
				h.logger.Warn("delete requests channel is full, processing immediately",
					zap.String("userID", userID),
					zap.Int("urlCount", len(ids)))

				h.processBatch(userID, ids)
			}(userID.(string), shortIDs)
		}

		c.JSON(http.StatusAccepted, gin.H{"message": "success"})
	}
}

// processDeleteRequests implements the fanIn pattern for processing delete requests
func (h *UserURLsDeleteHandler) processDeleteRequests() {
	defer h.wg.Done()

	// create a map to combine requests by user
	batches := make(map[string][]string)
	lastFlush := time.Now()

	ticker := time.NewTicker(h.flushInterval / 5)
	defer ticker.Stop()

	for {
		select {
		case req, ok := <-h.deleteRequests:
			if !ok {
				// channel was closed, process remaining requests and stop
				h.flushAllBatches(batches)
				return
			}

			// add URL to the batch for the corresponding user
			batches[req.userID] = append(batches[req.userID], req.shortIDs...)

			// if for some user the batch is full, process it
			for userID, shortIDs := range batches {
				if len(shortIDs) >= h.batchSize {
					h.processBatch(userID, shortIDs)
					// remove processed batch
					delete(batches, userID)
					lastFlush = time.Now()
				}
			}

		case <-ticker.C:
			// if enough time has passed since the last flush, process all remaining batches
			if time.Since(lastFlush) >= h.flushInterval && len(batches) > 0 {
				h.flushAllBatches(batches)
				batches = make(map[string][]string)
				lastFlush = time.Now()
			}

		case <-h.done:
			// received signal about the need to stop
			h.logger.Info("Stopping URL delete processor")
			h.flushAllBatches(batches)
			return
		}
	}
}

// processBatch processes one batch of delete requests
func (h *UserURLsDeleteHandler) processBatch(userID string, shortIDs []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.repository.MarkDeletedBatch(ctx, userID, shortIDs)
	if err != nil {
		h.logger.Error("failed to mark URLs as deleted",
			zap.String("userID", userID),
			zap.Int("count", len(shortIDs)),
			zap.Error(err))
	} else {
		h.logger.Info("successfully marked URLs as deleted",
			zap.String("userID", userID),
			zap.Int("count", len(shortIDs)))
	}
}

// flushAllBatches processes all remaining batches
func (h *UserURLsDeleteHandler) flushAllBatches(batches map[string][]string) {
	for userID, shortIDs := range batches {
		if len(shortIDs) > 0 {
			h.processBatch(userID, shortIDs)
		}
	}
}

// Shutdown correctly stops the handler
func (h *UserURLsDeleteHandler) Shutdown() {
	h.logger.Info("Shutting down UserURLsDeleteHandler")
	close(h.done)
	h.wg.Wait()
	h.logger.Info("UserURLsDeleteHandler shutdown complete")
}
