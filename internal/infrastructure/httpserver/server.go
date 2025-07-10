package httpserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

// Constants for the Server.
const (
	defaultReadHeaderTimeout = 15 * time.Second // default read header timeout
	defaultReadTimeout       = 15 * time.Second // default read timeout
	defaultWriteTimeout      = 10 * time.Second // default write timeout
	defaultIdleTimeout       = 30 * time.Second // default idle timeout
	defaultAddress           = "localhost:8080" // default address
)

// options is the options for the Server.
type options struct {
	handler           http.Handler
	address           string
	tlsCertPath       string
	tlsKeyPath        string
	readHeaderTimeout time.Duration
	readTimeout       time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration
	enableHTTPS       bool
}

// Option is the option for the Server.
type Option func(options *options) error

// Server is the HTTP server.
type Server struct {
	httpServer *http.Server
}

// New creates a new Server.
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

	if options.enableHTTPS {
		cert, err := tls.LoadX509KeyPair(options.tlsCertPath, options.tlsKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
		}
		server.httpServer.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{
				cert,
			},
		}
	}

	return server, nil
}

// Address returns the address of the Server.
func (s *Server) Address() string {
	return s.httpServer.Addr
}

// Start starts the Server.
func (s *Server) Start() error {
	if s.httpServer.TLSConfig != nil {
		return s.httpServer.ListenAndServeTLS("", "")
	}

	return s.httpServer.ListenAndServe()
}

// Shutdown shuts down the Server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// WithHandler is the option for the Server to set the handler.
func WithHandler(handler http.Handler) Option {
	return func(options *options) error {
		options.handler = handler
		return nil
	}
}

// WithAddress is the option for the Server to set the address.
func WithAddress(address string) Option {
	return func(options *options) error {
		options.address = address
		return nil
	}
}

// WithReadHeaderTimeout is the option for the Server to set the read header timeout.
func WithReadHeaderTimeout(readHeaderTimeout time.Duration) Option {
	return func(options *options) error {
		options.readHeaderTimeout = readHeaderTimeout
		return nil
	}
}

// WithReadTimeout is the option for the Server to set the read timeout.
func WithReadTimeout(readTimeout time.Duration) Option {
	return func(options *options) error {
		options.readTimeout = readTimeout
		return nil
	}
}

// WithWriteTimeout is the option for the Server to set the write timeout.
func WithWriteTimeout(writeTimeout time.Duration) Option {
	return func(options *options) error {
		options.writeTimeout = writeTimeout
		return nil
	}
}

// WithIdleTimeout is the option for the Server to set the idle timeout.
func WithIdleTimeout(idleTimeout time.Duration) Option {
	return func(options *options) error {
		options.idleTimeout = idleTimeout
		return nil
	}
}

// WithHTTPS is the option for the Server to set the enable HTTPS.
func WithHTTPS(enableHTTPS bool) Option {
	return func(options *options) error {
		options.enableHTTPS = enableHTTPS
		return nil
	}
}

func WithTLSCertPath(tlsCertPath string) Option {
	return func(options *options) error {
		options.tlsCertPath = tlsCertPath
		return nil
	}
}

func WithTLSKeyPath(tlsKeyPath string) Option {
	return func(options *options) error {
		options.tlsKeyPath = tlsKeyPath
		return nil
	}
}
