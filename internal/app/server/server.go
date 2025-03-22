package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type IHandler interface {
	IPattern
	IMethod
	IHandlerFunc
}

type IPattern interface {
	Pattern() string
}

type IMethod interface {
	Method() string
}

type IHandlerFunc interface {
	Handler() gin.HandlerFunc
}

type Server struct {
	httpServer *http.Server
	router     *gin.Engine
	handlers   []IHandler
}
type options struct {
	ginMode       string
	serverAddress string
	baseURL       string
	handlers      []IHandler
}
type Option func(options *options) error

func WithMode(mode string) Option {
	return func(o *options) error {
		o.ginMode = mode
		return nil
	}
}

func WithServerAddress(address string) Option {
	return func(o *options) error {
		o.serverAddress = address
		return nil
	}
}

func WithBaseURL(baseURL string) Option {
	return func(o *options) error {
		o.baseURL = baseURL
		return nil
	}
}

func WithHandler(handler IHandler) Option {
	return func(o *options) error {
		o.handlers = append(o.handlers, handler)
		return nil
	}
}

func NewServer(opts ...Option) (*Server, error) {
	options := &options{}
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, err
		}
	}

	// Set Gin mode
	switch options.ginMode {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		return nil, fmt.Errorf("invalid release mode: %s", options.ginMode)
	}

	// Create router and setup middleware
	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Setup handlers
	for _, handler := range options.handlers {
		router.Handle(handler.Method(), handler.Pattern(), handler.Handler())
	}

	return &Server{
		httpServer: &http.Server{
			Addr:              options.serverAddress,
			Handler:           router,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       60 * time.Second,
			ReadHeaderTimeout: 10 * time.Second,
		},
		router:   router,
		handlers: options.handlers,
	}, nil
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}
