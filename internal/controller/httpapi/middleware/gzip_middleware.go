package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
	"sync"
)

var (
	gzipWriterPool = sync.Pool{ //nolint:gochecknoglobals // gzip writer pool is used for gzip compression
		New: func() any {
			return gzip.NewWriter(nil)
		},
	}
	gzipReaderPool = sync.Pool{ //nolint:gochecknoglobals // gzip reader pool is used for gzip compression
		New: func() any {
			return new(gzip.Reader)
		},
	}
)

type gzipWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func GzipMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// request body is gzip
			if r.Header.Get("Content-Encoding") == "gzip" {
				reader, ok := gzipReaderPool.Get().(*gzip.Reader)
				if !ok {
					http.Error(w, "Failed to get gzip reader", http.StatusInternalServerError)
					return
				}
				if err := reader.Reset(r.Body); err != nil {
					gzipReaderPool.Put(reader)
					http.Error(w, "Failed to read gzipped request body", http.StatusBadRequest)
					return
				}

				r.Body = reader
				defer func() {
					_ = reader.Close()
					gzipReaderPool.Put(reader)
				}()
			}

			// response body is gzip
			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				writer, ok := gzipWriterPool.Get().(*gzip.Writer)
				if !ok {
					http.Error(w, "Failed to get gzip writer", http.StatusInternalServerError)
					return
				}
				writer.Reset(w)

				gzipWriter := &gzipWriter{
					ResponseWriter: w,
					writer:         writer,
				}

				w = gzipWriter
				w.Header().Set("Content-Encoding", "gzip")
				defer func() {
					_ = writer.Close()
					gzipWriterPool.Put(writer)
				}()
			}

			next.ServeHTTP(w, r)
		})
	}
}
