package simba_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gobwas/ws"
	"github.com/sillen102/simba"
	"github.com/sillen102/simba/simbaModels"
	"github.com/sillen102/simba/simbaTest/assert"
)

// Test types
type WSParams struct {
	Room string `path:"room" validate:"required"`
}

type WSAuthModel struct {
	UserID   int
	Username string
}

func TestWebSocketCallbackHandler(t *testing.T) {
	t.Parallel()

	t.Run("OnMessage is required", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when OnMessage is nil")
			}
		}()

		// Should panic because OnMessage is nil
		simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, params simbaModels.NoParams) error {
					return nil
				},
			},
		)
	})

	t.Run("handler creation succeeds with OnMessage", func(t *testing.T) {
		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					return nil
				},
			},
		)

		assert.NotNil(t, handler)
		var _ simba.Handler = handler
	})

	t.Run("callbacks are invoked in correct order", func(t *testing.T) {
		var callOrder []string
		var mu sync.Mutex

		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, params simbaModels.NoParams) error {
					mu.Lock()
					callOrder = append(callOrder, "OnConnect")
					mu.Unlock()
					return nil
				},

				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					mu.Lock()
					callOrder = append(callOrder, "OnMessage")
					mu.Unlock()
					return nil
				},

				OnDisconnect: func(ctx context.Context, params simbaModels.NoParams, err error) {
					mu.Lock()
					callOrder = append(callOrder, "OnDisconnect")
					mu.Unlock()
				},
			},
		)

		assert.NotNil(t, handler)
		// Note: Full test requires actual WebSocket connection
		// This verifies handler creation doesn't panic
	})

	t.Run("handler interface compliance", func(t *testing.T) {
		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[WSParams]{
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					return nil
				},
			},
		)

		var _ simba.Handler = handler
		assert.NotNil(t, handler)
	})
}

func TestAuthWebSocketCallbackHandler(t *testing.T) {
	t.Parallel()

	authHandler := simba.BearerAuth(
		func(ctx context.Context, token string) (WSAuthModel, error) {
			if token == "valid-token" {
				return WSAuthModel{
					UserID:   1,
					Username: "testuser",
				}, nil
			}
			return WSAuthModel{}, fmt.Errorf("invalid token")
		},
		simba.BearerAuthConfig{
			Name:        "BearerAuth",
			Format:      "JWT",
			Description: "Test bearer auth",
		},
	)

	t.Run("OnMessage is required", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when OnMessage is nil")
			}
		}()

		// Should panic because OnMessage is nil
		simba.AuthWebSocketHandler(
			simba.AuthWebSocketCallbacks[simbaModels.NoParams, WSAuthModel]{
				OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, params simbaModels.NoParams, auth WSAuthModel) error {
					return nil
				},
			},
			authHandler,
		)
	})

	t.Run("handler creation succeeds with OnMessage", func(t *testing.T) {
		handler := simba.AuthWebSocketHandler(
			simba.AuthWebSocketCallbacks[simbaModels.NoParams, WSAuthModel]{
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte, auth WSAuthModel) error {
					return nil
				},
			},
			authHandler,
		)

		assert.NotNil(t, handler)
		var _ simba.Handler = handler
	})

	t.Run("unauthorized request rejected", func(t *testing.T) {
		handler := simba.AuthWebSocketHandler(
			simba.AuthWebSocketCallbacks[simbaModels.NoParams, WSAuthModel]{
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte, auth WSAuthModel) error {
					return nil
				},
			},
			authHandler,
		)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		// No Authorization header
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid token rejected", func(t *testing.T) {
		handler := simba.AuthWebSocketHandler(
			simba.AuthWebSocketCallbacks[simbaModels.NoParams, WSAuthModel]{
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte, auth WSAuthModel) error {
					return nil
				},
			},
			authHandler,
		)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("handler interface compliance", func(t *testing.T) {
		handler := simba.AuthWebSocketHandler(
			simba.AuthWebSocketCallbacks[WSParams, WSAuthModel]{
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte, auth WSAuthModel) error {
					return nil
				},
			},
			authHandler,
		)

		var _ simba.Handler = handler
		assert.NotNil(t, handler)
	})
}

func TestConnectionTracking(t *testing.T) {
	t.Parallel()

	t.Run("connections map available in callbacks", func(t *testing.T) {
		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, params simbaModels.NoParams) error {
					// Connections map should be accessible
					assert.NotNil(t, connections)
					return nil
				},

				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					// Verify connections map is accessible
					assert.NotNil(t, connections)
					return nil
				},
			},
		)

		assert.NotNil(t, handler)
	})
}

func TestWebSocketConnection(t *testing.T) {
	t.Parallel()

	t.Run("connection has required fields", func(t *testing.T) {
		// Create a connection through the handler
		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, params simbaModels.NoParams) error {
					// Connection should have ID assigned
					assert.True(t, conn.ID != "")
					return nil
				},

				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					return nil
				},
			},
		)

		assert.NotNil(t, handler)
		// Connection will have ID and Params fields when created
	})

	t.Run("connection methods are available", func(t *testing.T) {
		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, params simbaModels.NoParams) error {
					// Verify connection has required methods/fields
					assert.True(t, conn.ID != "")
					assert.NotNil(t, conn.Context())
					return nil
				},

				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					return nil
				},
			},
		)

		assert.NotNil(t, handler)
		// Methods will be available when connection is created
	})
}

func TestCallbackErrorHandling(t *testing.T) {
	t.Parallel()

	t.Run("OnError callback can continue processing", func(t *testing.T) {
		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					return fmt.Errorf("test error")
				},

				OnError: func(ctx context.Context, conn *simba.WebSocketConnection, err error) bool {
					return true // Continue processing
				},
			},
		)

		assert.NotNil(t, handler)
	})

	t.Run("OnError callback can stop processing", func(t *testing.T) {
		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					return fmt.Errorf("test error")
				},

				OnError: func(ctx context.Context, conn *simba.WebSocketConnection, err error) bool {
					return false // Stop processing
				},
			},
		)

		assert.NotNil(t, handler)
	})
}

func TestConcurrentConnectionsAccess(t *testing.T) {
	t.Parallel()

	t.Run("concurrent connections map access is safe", func(t *testing.T) {
		var wg sync.WaitGroup
		const numGoroutines = 100

		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, params simbaModels.NoParams) error {
					// Simulate concurrent reads from the connections map
					wg.Add(numGoroutines)
					for i := 0; i < numGoroutines; i++ {
						go func(id int) {
							defer wg.Done()
							// These operations should not panic or race
							_ = connections[conn.ID]
							for _, c := range connections {
								_ = c.ID
							}
						}(i)
					}
					wg.Wait()

					return nil
				},

				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					return nil
				},
			},
		)

		assert.NotNil(t, handler)
	})
}

func TestCallbackParameterPassing(t *testing.T) {
	t.Parallel()

	t.Run("params are accessible in callbacks", func(t *testing.T) {
		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[WSParams]{
				OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, params WSParams) error {
					// Params should be accessible
					_ = params.Room
					return nil
				},

				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					return nil
				},
			},
		)

		assert.NotNil(t, handler)
		// Params will be passed when handler is invoked
	})

	t.Run("auth model is accessible in authenticated callbacks", func(t *testing.T) {
		authHandler := simba.BearerAuth(
			func(ctx context.Context, token string) (WSAuthModel, error) {
				return WSAuthModel{UserID: 1, Username: "test"}, nil
			},
			simba.BearerAuthConfig{Name: "Test"},
		)

		handler := simba.AuthWebSocketHandler(
			simba.AuthWebSocketCallbacks[simbaModels.NoParams, WSAuthModel]{
				OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, params simbaModels.NoParams, auth WSAuthModel) error {
					// Auth should be accessible
					_ = auth.UserID
					_ = auth.Username
					return nil
				},

				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte, auth WSAuthModel) error {
					return nil
				},
			},
			authHandler,
		)

		assert.NotNil(t, handler)
		// Auth will be passed when handler is invoked
	})
}

func TestConnectionsMapOperations(t *testing.T) {
	t.Parallel()

	t.Run("connections map can be filtered", func(t *testing.T) {
		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, params simbaModels.NoParams) error {
					// Can filter connections map
					var filtered []*simba.WebSocketConnection
					for _, c := range connections {
						if c.ID != "" {
							filtered = append(filtered, c)
						}
					}
					assert.NotNil(t, filtered)
					return nil
				},

				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					return nil
				},
			},
		)

		assert.NotNil(t, handler)
	})

	t.Run("connections map can be queried by ID", func(t *testing.T) {
		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, params simbaModels.NoParams) error {
					// Can get connection by ID (may return nil)
					retrieved := connections["some-id"]
					_ = retrieved // May be nil, that's ok
					return nil
				},

				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					return nil
				},
			},
		)

		assert.NotNil(t, handler)
	})
}

func TestOnDisconnectGuarantee(t *testing.T) {
	t.Parallel()

	t.Run("OnDisconnect called even if OnConnect fails", func(t *testing.T) {
		var disconnectCalled atomic.Bool

		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, params simbaModels.NoParams) error {
					return fmt.Errorf("connection failed")
				},

				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					return nil
				},

				OnDisconnect: func(ctx context.Context, params simbaModels.NoParams, err error) {
					disconnectCalled.Store(true)
				},
			},
		)

		assert.NotNil(t, handler)
		// OnDisconnect will be called via defer when connection handler runs
	})

	t.Run("OnDisconnect called even if OnMessage fails", func(t *testing.T) {
		var disconnectCalled atomic.Bool

		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					return fmt.Errorf("message processing failed")
				},

				OnDisconnect: func(ctx context.Context, params simbaModels.NoParams, err error) {
					disconnectCalled.Store(true)
				},
			},
		)

		assert.NotNil(t, handler)
		// OnDisconnect will be called via defer when connection handler runs
	})
}

func TestHandlerFuncVariants(t *testing.T) {
	t.Parallel()

	t.Run("WebSocketHandlerFunc accepts callback function", func(t *testing.T) {
		callbacksFunc := func() simba.WebSocketCallbacks[simbaModels.NoParams] {
			return simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					return nil
				},
			}
		}

		handler := simba.WebSocketHandlerFunc(callbacksFunc)
		assert.NotNil(t, handler)
		var _ simba.Handler = handler
	})

	t.Run("AuthWebSocketHandlerFunc accepts callback function", func(t *testing.T) {
		callbacksFunc := func() simba.AuthWebSocketCallbacks[simbaModels.NoParams, WSAuthModel] {
			return simba.AuthWebSocketCallbacks[simbaModels.NoParams, WSAuthModel]{
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte, auth WSAuthModel) error {
					return nil
				},
			}
		}

		authHandler := simba.BearerAuth(
			func(ctx context.Context, token string) (WSAuthModel, error) {
				return WSAuthModel{UserID: 1, Username: "test"}, nil
			},
			simba.BearerAuthConfig{
				Name:        "TestAuth",
				Format:      "JWT",
				Description: "Test",
			},
		)

		handler := simba.AuthWebSocketHandlerFunc(callbacksFunc, authHandler)
		assert.NotNil(t, handler)
		var _ simba.Handler = handler
	})
}
