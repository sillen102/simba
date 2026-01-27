package simba_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gobwas/ws/wsutil"
	"github.com/sillen102/simba"
	"github.com/sillen102/simba/simbaModels"
	"github.com/sillen102/simba/simbaTest/assert"
)

type WSParams struct {
	Room string `path:"room" validate:"required"`
	ID   string `query:"id"`
}

type WSAuthModel struct {
	UserID   int
	Username string
}

func TestWebSocketHandler(t *testing.T) {
	t.Parallel()

	t.Run("handler creation with params", func(t *testing.T) {
		handler := func(ctx context.Context, conn net.Conn, params WSParams) error {
			// Read one message and echo it back
			msg, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				return err
			}

			err = wsutil.WriteServerMessage(conn, op, msg)
			if err != nil {
				return err
			}

			return nil
		}

		h := simba.WebSocketHandler(handler)

		// Test that handler setup works (we can't easily test full WebSocket upgrade in unit tests)
		// For full integration testing, see the example
		var _ simba.Handler = h

		// Note: Full WebSocket upgrade testing requires a real server/client setup
		// This test verifies the handler implements the interface correctly
		assert.NotNil(t, h)
	})

	t.Run("validation error on missing required param", func(t *testing.T) {
		handler := func(ctx context.Context, conn net.Conn, params WSParams) error {
			return nil
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.SetPathValue("room", "") // Empty required field
		w := httptest.NewRecorder()

		h := simba.WebSocketHandler(handler)
		h.ServeHTTP(w, req)

		// Should get validation error before upgrade
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("handler type check", func(t *testing.T) {
		handler := func(ctx context.Context, conn net.Conn, params WSParams) error {
			return nil
		}

		h := simba.WebSocketHandler(handler)

		// Verify it implements the Handler interface
		var _ simba.Handler = h
	})
}

func TestAuthWebSocketHandler(t *testing.T) {
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

	t.Run("successful authenticated connection setup", func(t *testing.T) {
		handler := func(ctx context.Context, conn net.Conn, params WSParams, auth WSAuthModel) error {
			// Handler would process WebSocket messages here
			return nil
		}

		h := simba.AuthWebSocketHandler(handler, authHandler)

		// Verify it implements the Handler interface
		var _ simba.Handler = h

		// Note: Full WebSocket upgrade testing requires a real server/client setup
		// This test verifies the handler implements the interface correctly
		assert.NotNil(t, h)
	})

	t.Run("unauthorized connection fails", func(t *testing.T) {
		handler := func(ctx context.Context, conn net.Conn, params WSParams, auth WSAuthModel) error {
			return nil
		}

		req := httptest.NewRequest(http.MethodGet, "/chat", nil)
		req.SetPathValue("room", "chat")
		// No Authorization header
		w := httptest.NewRecorder()

		h := simba.AuthWebSocketHandler(handler, authHandler)
		h.ServeHTTP(w, req)

		// Should get 401 before upgrade
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid token fails", func(t *testing.T) {
		handler := func(ctx context.Context, conn net.Conn, params WSParams, auth WSAuthModel) error {
			return nil
		}

		req := httptest.NewRequest(http.MethodGet, "/chat", nil)
		req.SetPathValue("room", "chat")
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		h := simba.AuthWebSocketHandler(handler, authHandler)
		h.ServeHTTP(w, req)

		// Should get 401 before upgrade
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("handler type check", func(t *testing.T) {
		handler := func(ctx context.Context, conn net.Conn, params WSParams, auth WSAuthModel) error {
			return nil
		}

		h := simba.AuthWebSocketHandler(handler, authHandler)

		// Verify it implements the Handler interface
		var _ simba.Handler = h
	})
}

func TestWebSocketHandlerNoParams(t *testing.T) {
	t.Parallel()

	t.Run("websocket without params", func(t *testing.T) {
		handler := func(ctx context.Context, conn net.Conn, params simbaModels.NoParams) error {
			// Simple echo handler
			msg, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				return err
			}
			return wsutil.WriteServerMessage(conn, op, msg)
		}

		h := simba.WebSocketHandler(handler)

		// Verify it implements the Handler interface
		var _ simba.Handler = h
		assert.NotNil(t, h)
	})
}
