package server

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/cstone-io/twine/pkg/errors"
	"github.com/cstone-io/twine/pkg/logger"
)

// Server wraps an http.Server with graceful shutdown
type Server struct {
	Instance *http.Server
}

// NewServer creates a new Server with the given address and handler
func NewServer(addr string, handler http.Handler) *Server {
	if addr == "" {
		addr = ":3000"
	}

	return &Server{
		Instance: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}
}

// Start starts the server in a goroutine
func (s *Server) Start() {
	go func() {
		log := logger.Get()

		log.Info("Listening on %s", s.Instance.Addr)
		if err := s.Instance.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.CustomError(errors.ErrListenAndServe.Wrap(err))
		}
	}()
}

// AwaitShutdown blocks until context is cancelled, then gracefully shuts down
func (s *Server) AwaitShutdown(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.Instance.Shutdown(shutdownCtx); err != nil {
			logger.Get().CustomError(errors.ErrShutdownServer.Wrap(err))
		}
	}()
	wg.Wait()
	return nil
}
