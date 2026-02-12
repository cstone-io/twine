package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewServer tests server creation
func TestNewServer(t *testing.T) {
	t.Run("creates server with custom address", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		})

		srv := NewServer(":8080", handler)

		require.NotNil(t, srv)
		require.NotNil(t, srv.Instance)
		assert.Equal(t, ":8080", srv.Instance.Addr)
		assert.NotNil(t, srv.Instance.Handler)
	})

	t.Run("uses default address when empty", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		})

		srv := NewServer("", handler)

		require.NotNil(t, srv)
		assert.Equal(t, ":3000", srv.Instance.Addr)
	})

	t.Run("accepts different address formats", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

		testCases := []string{
			":8080",
			"localhost:8080",
			"127.0.0.1:8080",
			"0.0.0.0:8080",
		}

		for _, addr := range testCases {
			srv := NewServer(addr, handler)
			assert.Equal(t, addr, srv.Instance.Addr)
		}
	})

	t.Run("stores handler correctly", func(t *testing.T) {
		called := false

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		})

		srv := NewServer(":8080", handler)

		// Test that handler is stored correctly
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		srv.Instance.Handler.ServeHTTP(w, r)

		assert.True(t, called)
	})
}

// TestServer_Start tests server startup
func TestServer_Start(t *testing.T) {
	t.Run("starts server in goroutine", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		})

		// Use a random available port
		srv := NewServer(":0", handler)
		srv.Start()

		// Give it a moment to start
		time.Sleep(10 * time.Millisecond)

		// Server should be running in background
		// We can't easily test if it's listening without making a real connection
		// So we just verify Start() returns immediately (runs in goroutine)
	})

	t.Run("Start returns immediately", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

		srv := NewServer(":0", handler)

		start := time.Now()
		srv.Start()
		duration := time.Since(start)

		// Start should return almost immediately (< 10ms)
		assert.Less(t, duration, 10*time.Millisecond)
	})
}

// TestServer_AwaitShutdown tests graceful shutdown
func TestServer_AwaitShutdown(t *testing.T) {
	t.Run("blocks until context is cancelled", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

		srv := NewServer(":0", handler)
		srv.Start()

		ctx, cancel := context.WithCancel(context.Background())

		// Track if AwaitShutdown returned
		done := make(chan bool)
		go func() {
			err := srv.AwaitShutdown(ctx)
			assert.NoError(t, err)
			done <- true
		}()

		// Give it a moment
		time.Sleep(10 * time.Millisecond)

		// Should still be blocking
		select {
		case <-done:
			t.Fatal("AwaitShutdown returned before context was cancelled")
		default:
			// Good, still blocking
		}

		// Cancel context
		cancel()

		// Should return now
		select {
		case <-done:
			// Good, returned after cancel
		case <-time.After(2 * time.Second):
			t.Fatal("AwaitShutdown did not return after context cancel")
		}
	})

	t.Run("gracefully shuts down server", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

		srv := NewServer(":0", handler)
		srv.Start()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := srv.AwaitShutdown(ctx)
		assert.NoError(t, err)
	})

	t.Run("handles already cancelled context", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

		srv := NewServer(":0", handler)
		srv.Start()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := srv.AwaitShutdown(ctx)
		assert.NoError(t, err)
	})
}

// TestServer_Integration tests realistic server scenarios
func TestServer_Integration(t *testing.T) {
	t.Run("complete server lifecycle", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("ok"))
		})

		// Create and start server
		srv := NewServer(":0", handler)
		srv.Start()

		// Give it time to start
		time.Sleep(50 * time.Millisecond)

		// Shutdown gracefully
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err := srv.AwaitShutdown(ctx)
		assert.NoError(t, err)
	})

	t.Run("server with different handlers", func(t *testing.T) {
		testCases := []struct {
			name    string
			handler http.Handler
		}{
			{
				"simple handler",
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
				}),
			},
			{
				"mux handler",
				http.NewServeMux(),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				srv := NewServer(":0", tc.handler)
				assert.NotNil(t, srv)
				assert.NotNil(t, srv.Instance.Handler)
			})
		}
	})

	t.Run("multiple sequential start/stop cycles", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

		for i := 0; i < 3; i++ {
			srv := NewServer(":0", handler)
			srv.Start()

			time.Sleep(10 * time.Millisecond)

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			err := srv.AwaitShutdown(ctx)
			cancel()

			assert.NoError(t, err)
		}
	})
}

// TestServer_DefaultAddress tests default address handling
func TestServer_DefaultAddress(t *testing.T) {
	t.Run("empty string uses default port", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		srv := NewServer("", handler)
		assert.Equal(t, ":3000", srv.Instance.Addr)
	})

	t.Run("explicit port is preserved", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		srv := NewServer(":4000", handler)
		assert.Equal(t, ":4000", srv.Instance.Addr)
	})
}
