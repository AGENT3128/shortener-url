package httpserver

import (
	"context"
	"net/http"
	"time"
)

const (
	defaultReadHeaderTimeout = 15 * time.Second
	defaultReadTimeout       = 15 * time.Second
	defaultWriteTimeout      = 10 * time.Second
	defaultIdleTimeout       = 30 * time.Second
	defaultAddress           = "localhost:8080"
)

type options struct {
	address           string
	readHeaderTimeout time.Duration
	readTimeout       time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration
	handler           http.Handler
}

type Option func(options *options) error

type Server struct {
	httpServer *http.Server
}

func New(opts ...Option) (*Server, error) {
	server := &Server{
		httpServer: &http.Server{
			Addr:              defaultAddress,
			ReadHeaderTimeout: defaultReadHeaderTimeout,
			ReadTimeout:       defaultReadTimeout,
			WriteTimeout:      defaultWriteTimeout,
			IdleTimeout:       defaultIdleTimeout,
			Handler:           nil,
		},
	}
	options := &options{}
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, err
		}
	}

	if options.address != "" {
		server.httpServer.Addr = options.address
	}

	if options.readHeaderTimeout != 0 {
		server.httpServer.ReadHeaderTimeout = options.readHeaderTimeout
	}

	if options.readTimeout != 0 {
		server.httpServer.ReadTimeout = options.readTimeout
	}

	if options.writeTimeout != 0 {
		server.httpServer.WriteTimeout = options.writeTimeout
	}

	if options.idleTimeout != 0 {
		server.httpServer.IdleTimeout = options.idleTimeout
	}

	if options.handler != nil {
		server.httpServer.Handler = options.handler
	}

	return server, nil
}

func (s *Server) Address() string {
	return s.httpServer.Addr
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func WithHandler(handler http.Handler) Option {
	return func(options *options) error {
		options.handler = handler
		return nil
	}
}

func WithAddress(address string) Option {
	return func(options *options) error {
		options.address = address
		return nil
	}
}

func WithReadHeaderTimeout(readHeaderTimeout time.Duration) Option {
	return func(options *options) error {
		options.readHeaderTimeout = readHeaderTimeout
		return nil
	}
}

func WithReadTimeout(readTimeout time.Duration) Option {
	return func(options *options) error {
		options.readTimeout = readTimeout
		return nil
	}
}

func WithWriteTimeout(writeTimeout time.Duration) Option {
	return func(options *options) error {
		options.writeTimeout = writeTimeout
		return nil
	}
}

func WithIdleTimeout(idleTimeout time.Duration) Option {
	return func(options *options) error {
		options.idleTimeout = idleTimeout
		return nil
	}
}
