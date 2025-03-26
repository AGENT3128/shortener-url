package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	gzipWriterPool = sync.Pool{
		New: func() any {
			return gzip.NewWriter(nil)
		},
	}
	gzipReaderPool = sync.Pool{
		New: func() any {
			return new(gzip.Reader)
		},
	}
)

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func GzipMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		// request body is gzip
		if c.GetHeader("Content-Encoding") == "gzip" {
			reader := gzipReaderPool.Get().(*gzip.Reader)
			if err := reader.Reset(c.Request.Body); err != nil {
				gzipReaderPool.Put(reader)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read gzipped request body"})
				return
			}

			c.Request.Body = reader
			defer func() {
				reader.Close()
				gzipReaderPool.Put(reader)
			}()
		}

		// response body is gzip
		if strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {

			writer := gzipWriterPool.Get().(*gzip.Writer)
			writer.Reset(c.Writer)

			gzipWriter := &gzipWriter{
				ResponseWriter: c.Writer,
				writer:         writer,
			}

			c.Writer = gzipWriter
			c.Writer.Header().Set("Content-Encoding", "gzip")
			defer func() {
				writer.Close()
				gzipWriterPool.Put(writer)
			}()
		}

		c.Next()
	}
}
