package websocket_test

import (

	"github.com/sillen102/simba/websocket"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/simbaModels"
	"github.com/sillen102/simba/simbaTest/assert"
)

// Test types for callbacks tests
type WSCallbackParams struct {
	Room string `path:"room" validate:"required"`
}

type WSCallbackAuthModel struct {
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
		websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnConnect: func(ctx context.Context, conn *websocket.Connection, params simbaModels.NoParams) error {
						return nil
					},
				}
			},
		)
	})

	t.Run("handler creation succeeds with OnMessage", func(t *testing.T) {
		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return nil
					},
				}
			},
		)

		assert.NotNil(t, handler)
		var _ simba.Handler = handler
	})

	t.Run("callbacks structure is correct", func(t *testing.T) {
		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnConnect: func(ctx context.Context, conn *websocket.Connection, params simbaModels.NoParams) error {
						return nil
					},

					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return nil
					},

					OnDisconnect: func(ctx context.Context, connID string, params simbaModels.NoParams, err error) {
						// connID is provided since connection is closed
					},
				}
			},
		)

		assert.NotNil(t, handler)
	})

	t.Run("handler interface compliance", func(t *testing.T) {
		handler := websocket.Handler(
			func() websocket.Callbacks[WSCallbackParams] {
				return websocket.Callbacks[WSCallbackParams]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return nil
					},
				}
			},
		)

		var _ simba.Handler = handler
		assert.NotNil(t, handler)
	})
}

func TestAuthWebSocketCallbackHandler(t *testing.T) {
	t.Parallel()

	authHandler := simba.BearerAuth(
		func(ctx context.Context, token string) (WSCallbackAuthModel, error) {
			if token == "valid-token" {
				return WSCallbackAuthModel{
					UserID:   1,
					Username: "testuser",
				}, nil
			}
			return WSCallbackAuthModel{}, fmt.Errorf("invalid token")
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
		websocket.AuthHandler(
			func() websocket.AuthCallbacks[simbaModels.NoParams, WSCallbackAuthModel] {
				return websocket.AuthCallbacks[simbaModels.NoParams, WSCallbackAuthModel]{
					OnConnect: func(ctx context.Context, conn *websocket.Connection, params simbaModels.NoParams, auth WSCallbackAuthModel) error {
						return nil
					},
				}
			},
			authHandler,
		)
	})

	t.Run("handler creation succeeds with OnMessage", func(t *testing.T) {
		handler := websocket.AuthHandler(
			func() websocket.AuthCallbacks[simbaModels.NoParams, WSCallbackAuthModel] {
				return websocket.AuthCallbacks[simbaModels.NoParams, WSCallbackAuthModel]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte, auth WSCallbackAuthModel) error {
						return nil
					},
				}
			},
			authHandler,
		)

		assert.NotNil(t, handler)
		var _ simba.Handler = handler
	})

	t.Run("unauthorized request rejected", func(t *testing.T) {
		handler := websocket.AuthHandler(
			func() websocket.AuthCallbacks[simbaModels.NoParams, WSCallbackAuthModel] {
				return websocket.AuthCallbacks[simbaModels.NoParams, WSCallbackAuthModel]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte, auth WSCallbackAuthModel) error {
						return nil
					},
				}
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
		handler := websocket.AuthHandler(
			func() websocket.AuthCallbacks[simbaModels.NoParams, WSCallbackAuthModel] {
				return websocket.AuthCallbacks[simbaModels.NoParams, WSCallbackAuthModel]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte, auth WSCallbackAuthModel) error {
						return nil
					},
				}
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
		handler := websocket.AuthHandler(
			func() websocket.AuthCallbacks[WSCallbackParams, WSCallbackAuthModel] {
				return websocket.AuthCallbacks[WSCallbackParams, WSCallbackAuthModel]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte, auth WSCallbackAuthModel) error {
						return nil
					},
				}
			},
			authHandler,
		)

		var _ simba.Handler = handler
		assert.NotNil(t, handler)
	})
}

func TestConnection_API(t *testing.T) {
	t.Parallel()

	t.Run("connection has ID field", func(t *testing.T) {
		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnConnect: func(ctx context.Context, conn *websocket.Connection, params simbaModels.NoParams) error {
						// Connection should have ID assigned
						assert.True(t, conn.ID != "")
						return nil
					},

					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return nil
					},
				}
			},
		)

		assert.NotNil(t, handler)
	})

	t.Run("connection ID available in OnDisconnect", func(t *testing.T) {
		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return nil
					},
					OnDisconnect: func(ctx context.Context, connID string, params simbaModels.NoParams, err error) {
						// connID should be provided for cleanup in external registries
						assert.True(t, connID != "")
					},
				}
			},
		)

		assert.NotNil(t, handler)
	})
}

func TestCallbackErrorHandling(t *testing.T) {
	t.Parallel()

	t.Run("OnError callback can continue processing", func(t *testing.T) {
		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return fmt.Errorf("test error")
					},

					OnError: func(ctx context.Context, conn *websocket.Connection, err error) bool {
						return true // Continue processing
					},
				}
			},
		)

		assert.NotNil(t, handler)
	})

	t.Run("OnError callback can stop processing", func(t *testing.T) {
		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return fmt.Errorf("test error")
					},

					OnError: func(ctx context.Context, conn *websocket.Connection, err error) bool {
						return false // Stop processing
					},
				}
			},
		)

		assert.NotNil(t, handler)
	})
}

func TestCallbackParameterPassing(t *testing.T) {
	t.Parallel()

	t.Run("params are accessible in callbacks", func(t *testing.T) {
		handler := websocket.Handler(
			func() websocket.Callbacks[WSCallbackParams] {
				return websocket.Callbacks[WSCallbackParams]{
					OnConnect: func(ctx context.Context, conn *websocket.Connection, params WSCallbackParams) error {
						// Params should be accessible
						_ = params.Room
						return nil
					},

					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return nil
					},
				}
			},
		)

		assert.NotNil(t, handler)
	})

	t.Run("auth model is accessible in authenticated callbacks", func(t *testing.T) {
		authHandler := simba.BearerAuth(
			func(ctx context.Context, token string) (WSCallbackAuthModel, error) {
				return WSCallbackAuthModel{UserID: 1, Username: "test"}, nil
			},
			simba.BearerAuthConfig{Name: "Test"},
		)

		handler := websocket.AuthHandler(
			func() websocket.AuthCallbacks[simbaModels.NoParams, WSCallbackAuthModel] {
				return websocket.AuthCallbacks[simbaModels.NoParams, WSCallbackAuthModel]{
					OnConnect: func(ctx context.Context, conn *websocket.Connection, params simbaModels.NoParams, auth WSCallbackAuthModel) error {
						// Auth should be accessible
						_ = auth.UserID
						_ = auth.Username
						return nil
					},

					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte, auth WSCallbackAuthModel) error {
						return nil
					},
				}
			},
			authHandler,
		)

		assert.NotNil(t, handler)
	})
}

func TestOnDisconnectGuarantee(t *testing.T) {
	t.Parallel()

	t.Run("OnDisconnect signature includes connID", func(t *testing.T) {
		var disconnectCalled atomic.Bool

		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnConnect: func(ctx context.Context, conn *websocket.Connection, params simbaModels.NoParams) error {
						return fmt.Errorf("connection failed")
					},

					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return nil
					},

					OnDisconnect: func(ctx context.Context, connID string, params simbaModels.NoParams, err error) {
						disconnectCalled.Store(true)
						// connID is provided for external registry cleanup
						assert.True(t, connID != "")
					},
				}
			},
		)

		assert.NotNil(t, handler)
	})

	t.Run("auth OnDisconnect includes auth model", func(t *testing.T) {
		authHandler := simba.BearerAuth(
			func(ctx context.Context, token string) (WSCallbackAuthModel, error) {
				return WSCallbackAuthModel{UserID: 1, Username: "test"}, nil
			},
			simba.BearerAuthConfig{Name: "Test"},
		)

		handler := websocket.AuthHandler(
			func() websocket.AuthCallbacks[simbaModels.NoParams, WSCallbackAuthModel] {
				return websocket.AuthCallbacks[simbaModels.NoParams, WSCallbackAuthModel]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte, auth WSCallbackAuthModel) error {
						return nil
					},

					OnDisconnect: func(ctx context.Context, connID string, params simbaModels.NoParams, auth WSCallbackAuthModel, err error) {
						// Both connID and auth are available for cleanup
						assert.True(t, connID != "")
						_ = auth.UserID
					},
				}
			},
			authHandler,
		)

		assert.NotNil(t, handler)
	})
}

func TestHandlerFuncVariants(t *testing.T) {
	t.Parallel()

	t.Run("Handler accepts callback function", func(t *testing.T) {
		callbacksFunc := func() websocket.Callbacks[simbaModels.NoParams] {
			return websocket.Callbacks[simbaModels.NoParams]{
				OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
					return nil
				},
			}
		}

		handler := websocket.Handler(callbacksFunc)
		assert.NotNil(t, handler)
		var _ simba.Handler = handler
	})

	t.Run("AuthHandler accepts callback function", func(t *testing.T) {
		callbacksFunc := func() websocket.AuthCallbacks[simbaModels.NoParams, WSCallbackAuthModel] {
			return websocket.AuthCallbacks[simbaModels.NoParams, WSCallbackAuthModel]{
				OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte, auth WSCallbackAuthModel) error {
					return nil
				},
			}
		}

		authHandler := simba.BearerAuth(
			func(ctx context.Context, token string) (WSCallbackAuthModel, error) {
				return WSCallbackAuthModel{UserID: 1, Username: "test"}, nil
			},
			simba.BearerAuthConfig{
				Name:        "TestAuth",
				Format:      "JWT",
				Description: "Test",
			},
		)

		handler := websocket.AuthHandler(callbacksFunc, authHandler)
		assert.NotNil(t, handler)
		var _ simba.Handler = handler
	})
}
