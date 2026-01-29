package websocket_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sillen102/simba/websocket"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/sillen102/simba"
	"github.com/sillen102/simba/simbaContext"
	"github.com/sillen102/simba/simbaModels"
	"github.com/sillen102/simba/simbaTest/assert"
)

func TestHandler_ConnectionLifecycle(t *testing.T) {
	t.Parallel()

	t.Run("OnConnect is called on connection", func(t *testing.T) {
		var connectCalled atomic.Bool
		var receivedConnID string

		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnConnect: func(ctx context.Context, conn *websocket.Connection, params simbaModels.NoParams) error {
						connectCalled.Store(true)
						receivedConnID = conn.ID
						assert.NotNil(t, conn)
						assert.True(t, conn.ID != "")
						return nil
					},
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return nil
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		time.Sleep(50 * time.Millisecond)

		assert.True(t, connectCalled.Load())
		assert.True(t, receivedConnID != "")
	})

	t.Run("OnMessage is called when message received", func(t *testing.T) {
		var messageCalled atomic.Bool
		var receivedData string

		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						messageCalled.Store(true)
						receivedData = string(data)
						return nil
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		testMessage := "Hello WebSocket"
		err = wsutil.WriteClientText(conn, []byte(testMessage))
		assert.NoError(t, err)

		time.Sleep(50 * time.Millisecond)

		assert.True(t, messageCalled.Load())
		assert.Equal(t, testMessage, receivedData)
	})

	t.Run("OnMessage receives binary messages", func(t *testing.T) {
		var receivedData []byte

		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						receivedData = data
						return nil
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		testData := []byte{0x01, 0x02, 0x03, 0x04}
		err = wsutil.WriteClientBinary(conn, testData)
		assert.NoError(t, err)

		time.Sleep(50 * time.Millisecond)

		assert.Equal(t, testData, receivedData)
	})

	t.Run("OnDisconnect is called on connection close", func(t *testing.T) {
		var disconnectCalled atomic.Bool
		var disconnectConnID string

		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return nil
					},
					OnDisconnect: func(ctx context.Context, connID string, params simbaModels.NoParams, err error) {
						disconnectCalled.Store(true)
						disconnectConnID = connID
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)

		conn.Close()

		time.Sleep(50 * time.Millisecond)

		assert.True(t, disconnectCalled.Load())
		assert.True(t, disconnectConnID != "")
	})

	t.Run("OnDisconnect receives connection ID", func(t *testing.T) {
		var connectConnID string
		var disconnectConnID string

		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnConnect: func(ctx context.Context, conn *websocket.Connection, params simbaModels.NoParams) error {
						connectConnID = conn.ID
						return nil
					},
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return nil
					},
					OnDisconnect: func(ctx context.Context, connID string, params simbaModels.NoParams, err error) {
						disconnectConnID = connID
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)

		time.Sleep(50 * time.Millisecond)
		conn.Close()
		time.Sleep(50 * time.Millisecond)

		assert.True(t, connectConnID != "")
		assert.Equal(t, connectConnID, disconnectConnID)
	})

	t.Run("OnError is called on message error", func(t *testing.T) {
		var errorCalled atomic.Bool

		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return fmt.Errorf("test error")
					},
					OnError: func(ctx context.Context, conn *websocket.Connection, err error) bool {
						errorCalled.Store(true)
						return false
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		err = wsutil.WriteClientText(conn, []byte("test"))
		assert.NoError(t, err)

		time.Sleep(50 * time.Millisecond)

		assert.True(t, errorCalled.Load())
	})
}

func TestHandler_ConnectionID(t *testing.T) {
	t.Parallel()

	t.Run("each connection gets unique ID", func(t *testing.T) {
		var mu sync.Mutex
		var connIDs []string

		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnConnect: func(ctx context.Context, conn *websocket.Connection, params simbaModels.NoParams) error {
						mu.Lock()
						defer mu.Unlock()
						connIDs = append(connIDs, conn.ID)
						return nil
					},
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return nil
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn1, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn1.Close()

		conn2, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn2.Close()

		conn3, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn3.Close()

		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		defer mu.Unlock()

		assert.Equal(t, 3, len(connIDs))
		seen := make(map[string]bool)
		for _, id := range connIDs {
			assert.False(t, seen[id], "duplicate connection ID")
			seen[id] = true
		}
	})

	t.Run("connection ID is UUID format", func(t *testing.T) {
		var connID string

		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnConnect: func(ctx context.Context, conn *websocket.Connection, params simbaModels.NoParams) error {
						connID = conn.ID
						return nil
					},
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return nil
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		time.Sleep(50 * time.Millisecond)

		assert.Equal(t, 36, len(connID))
		assert.Equal(t, '-', rune(connID[8]))
		assert.Equal(t, '-', rune(connID[13]))
		assert.Equal(t, '-', rune(connID[18]))
		assert.Equal(t, '-', rune(connID[23]))
	})
}

func TestHandler_WriteOperations(t *testing.T) {
	t.Parallel()

	t.Run("WriteText sends text message", func(t *testing.T) {
		// Echo back text messages
		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return conn.WriteText("echo: " + string(data))
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		// Send a message to trigger the response
		err = wsutil.WriteClientText(conn, []byte("hello"))
		assert.NoError(t, err)

		conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		msg, op, err := wsutil.ReadServerData(conn)
		assert.NoError(t, err)
		assert.Equal(t, ws.OpText, op)
		assert.Equal(t, "echo: hello", string(msg))
	})

	t.Run("WriteBinary sends binary message", func(t *testing.T) {
		expectedData := []byte{0xDE, 0xAD, 0xBE, 0xEF}

		// Echo back binary data
		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return conn.WriteBinary(expectedData)
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		// Send a message to trigger the response
		err = wsutil.WriteClientText(conn, []byte("send binary"))
		assert.NoError(t, err)

		conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		msg, op, err := wsutil.ReadServerData(conn)
		assert.NoError(t, err)
		assert.Equal(t, ws.OpBinary, op)
		assert.Equal(t, expectedData, msg)
	})

	t.Run("WriteJSON sends JSON message", func(t *testing.T) {
		type TestMessage struct {
			Type string `json:"type"`
			Data string `json:"data"`
		}

		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return conn.WriteJSON(TestMessage{Type: "response", Data: string(data)})
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		// Send a message to trigger the response
		err = wsutil.WriteClientText(conn, []byte("hello"))
		assert.NoError(t, err)

		conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		msg, op, err := wsutil.ReadServerData(conn)
		assert.NoError(t, err)
		assert.Equal(t, ws.OpText, op)
		assert.Equal(t, `{"type":"response","data":"hello"}`, string(msg))
	})
}

func TestHandler_ExternalRegistry(t *testing.T) {
	t.Parallel()

	t.Run("can implement external connection registry", func(t *testing.T) {
		registry := &sync.Map{}

		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnConnect: func(ctx context.Context, conn *websocket.Connection, params simbaModels.NoParams) error {
						registry.Store(conn.ID, conn)
						return nil
					},
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return nil
					},
					OnDisconnect: func(ctx context.Context, connID string, params simbaModels.NoParams, err error) {
						registry.Delete(connID)
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)

		time.Sleep(50 * time.Millisecond)

		count := 0
		registry.Range(func(key, value any) bool {
			count++
			return true
		})
		assert.Equal(t, 1, count)

		conn.Close()
		time.Sleep(50 * time.Millisecond)

		count = 0
		registry.Range(func(key, value any) bool {
			count++
			return true
		})
		assert.Equal(t, 0, count)
	})

	t.Run("can send to connection from external source", func(t *testing.T) {
		registry := &sync.Map{}
		var registeredConnID string

		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnConnect: func(ctx context.Context, conn *websocket.Connection, params simbaModels.NoParams) error {
						registeredConnID = conn.ID
						registry.Store(conn.ID, conn)
						return nil
					},
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return nil
					},
					OnDisconnect: func(ctx context.Context, connID string, params simbaModels.NoParams, err error) {
						registry.Delete(connID)
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		time.Sleep(50 * time.Millisecond)

		conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		if wsConn, ok := registry.Load(registeredConnID); ok {
			wsConn.(*websocket.Connection).WriteText("external message")
		}

		msg, _, err := wsutil.ReadServerData(conn)
		assert.NoError(t, err)
		assert.Equal(t, "external message", string(msg))
	})
}

func TestAuthHandler_Authentication(t *testing.T) {
	t.Parallel()

	authHandler := simba.BearerAuth(
		func(ctx context.Context, token string) (WSAuthModel, error) {
			if token == "valid-token" {
				return WSAuthModel{
					UserID:   123,
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

	t.Run("authenticated connection succeeds with valid token", func(t *testing.T) {
		var authReceived WSAuthModel

		handler := websocket.AuthHandler(
			func() websocket.AuthCallbacks[simbaModels.NoParams, WSAuthModel] {
				return websocket.AuthCallbacks[simbaModels.NoParams, WSAuthModel]{
					OnConnect: func(ctx context.Context, conn *websocket.Connection, params simbaModels.NoParams, auth WSAuthModel) error {
						authReceived = auth
						return nil
					},
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte, auth WSAuthModel) error {
						return nil
					},
				}
			},
			authHandler,
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		dialer := ws.Dialer{
			Header: ws.HandshakeHeaderHTTP(http.Header{
				"Authorization": []string{"Bearer valid-token"},
			}),
		}
		conn, _, _, err := dialer.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		time.Sleep(50 * time.Millisecond)

		assert.Equal(t, 123, authReceived.UserID)
		assert.Equal(t, "testuser", authReceived.Username)
	})

	t.Run("connection rejected with invalid token", func(t *testing.T) {
		handler := websocket.AuthHandler(
			func() websocket.AuthCallbacks[simbaModels.NoParams, WSAuthModel] {
				return websocket.AuthCallbacks[simbaModels.NoParams, WSAuthModel]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte, auth WSAuthModel) error {
						return nil
					},
				}
			},
			authHandler,
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		dialer := ws.Dialer{
			Header: ws.HandshakeHeaderHTTP(http.Header{
				"Authorization": []string{"Bearer invalid-token"},
			}),
		}
		_, _, _, err := dialer.Dial(context.Background(), "ws"+server.URL[4:])
		assert.Error(t, err)
	})

	t.Run("auth is passed to OnMessage", func(t *testing.T) {
		var messageAuth WSAuthModel

		handler := websocket.AuthHandler(
			func() websocket.AuthCallbacks[simbaModels.NoParams, WSAuthModel] {
				return websocket.AuthCallbacks[simbaModels.NoParams, WSAuthModel]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte, auth WSAuthModel) error {
						messageAuth = auth
						return nil
					},
				}
			},
			authHandler,
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		dialer := ws.Dialer{
			Header: ws.HandshakeHeaderHTTP(http.Header{
				"Authorization": []string{"Bearer valid-token"},
			}),
		}
		conn, _, _, err := dialer.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		err = wsutil.WriteClientText(conn, []byte("test"))
		assert.NoError(t, err)

		time.Sleep(50 * time.Millisecond)

		assert.Equal(t, 123, messageAuth.UserID)
		assert.Equal(t, "testuser", messageAuth.Username)
	})

	t.Run("auth is passed to OnDisconnect", func(t *testing.T) {
		var disconnectAuth WSAuthModel
		var disconnectConnID string

		handler := websocket.AuthHandler(
			func() websocket.AuthCallbacks[simbaModels.NoParams, WSAuthModel] {
				return websocket.AuthCallbacks[simbaModels.NoParams, WSAuthModel]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte, auth WSAuthModel) error {
						return nil
					},
					OnDisconnect: func(ctx context.Context, connID string, params simbaModels.NoParams, auth WSAuthModel, err error) {
						disconnectConnID = connID
						disconnectAuth = auth
					},
				}
			},
			authHandler,
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		dialer := ws.Dialer{
			Header: ws.HandshakeHeaderHTTP(http.Header{
				"Authorization": []string{"Bearer valid-token"},
			}),
		}
		conn, _, _, err := dialer.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)

		time.Sleep(50 * time.Millisecond)
		conn.Close()
		time.Sleep(50 * time.Millisecond)

		assert.True(t, disconnectConnID != "")
		assert.Equal(t, 123, disconnectAuth.UserID)
		assert.Equal(t, "testuser", disconnectAuth.Username)
	})
}

func TestHandler_ConcurrentConnections(t *testing.T) {
	t.Parallel()

	t.Run("handles multiple concurrent connections", func(t *testing.T) {
		var connectionCount atomic.Int32

		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnConnect: func(ctx context.Context, conn *websocket.Connection, params simbaModels.NoParams) error {
						connectionCount.Add(1)
						return nil
					},
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return nil
					},
					OnDisconnect: func(ctx context.Context, connID string, params simbaModels.NoParams, err error) {
						connectionCount.Add(-1)
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		numClients := 10
		var wg sync.WaitGroup
		wg.Add(numClients)

		conns := make([]interface{ Close() error }, numClients)
		var connsMu sync.Mutex

		for i := 0; i < numClients; i++ {
			go func(idx int) {
				defer wg.Done()
				conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
				if err == nil {
					connsMu.Lock()
					conns[idx] = conn
					connsMu.Unlock()
				}
			}(i)
		}

		wg.Wait()
		time.Sleep(100 * time.Millisecond)

		assert.Equal(t, int32(numClients), connectionCount.Load())

		for _, conn := range conns {
			if conn != nil {
				conn.Close()
			}
		}

		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, int32(0), connectionCount.Load())
	})
}

func TestHandler_ErrorRecovery(t *testing.T) {
	t.Parallel()

	t.Run("OnError can continue processing after error", func(t *testing.T) {
		var messageCount atomic.Int32

		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						messageCount.Add(1)
						if messageCount.Load() == 1 {
							return fmt.Errorf("first message error")
						}
						return nil
					},
					OnError: func(ctx context.Context, conn *websocket.Connection, err error) bool {
						return true
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		err = wsutil.WriteClientText(conn, []byte("message1"))
		assert.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		err = wsutil.WriteClientText(conn, []byte("message2"))
		assert.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		assert.Equal(t, int32(2), messageCount.Load())
	})

	t.Run("connection closes when OnError returns false", func(t *testing.T) {
		var disconnectCalled atomic.Bool

		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return fmt.Errorf("error")
					},
					OnError: func(ctx context.Context, conn *websocket.Connection, err error) bool {
						return false
					},
					OnDisconnect: func(ctx context.Context, connID string, params simbaModels.NoParams, err error) {
						disconnectCalled.Store(true)
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		err = wsutil.WriteClientText(conn, []byte("test"))
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		assert.True(t, disconnectCalled.Load())
	})
}

func TestHandler_ThreadSafety(t *testing.T) {
	t.Parallel()

	t.Run("concurrent writes are thread-safe", func(t *testing.T) {
		var wsConn *websocket.Connection
		var connReady sync.WaitGroup
		connReady.Add(1)

		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnConnect: func(ctx context.Context, conn *websocket.Connection, params simbaModels.NoParams) error {
						wsConn = conn
						connReady.Done()
						return nil
					},
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return nil
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		connReady.Wait()

		var wg sync.WaitGroup
		numWriters := 10
		messagesPerWriter := 100

		for i := 0; i < numWriters; i++ {
			wg.Add(1)
			go func(writerID int) {
				defer wg.Done()
				for j := 0; j < messagesPerWriter; j++ {
					wsConn.WriteText(fmt.Sprintf("writer-%d-msg-%d", writerID, j))
				}
			}(i)
		}

		wg.Wait()
	})
}

// WSAuthModel is a test auth model
type WSAuthModel struct {
	UserID   int
	Username string
}

func TestHandler_Middleware(t *testing.T) {
	t.Parallel()

	t.Run("middleware runs before callbacks", func(t *testing.T) {
		var middlewareCalled atomic.Bool
		var callbackCalled atomic.Bool
		var middlewareBeforeCallback atomic.Bool

		testMiddleware := func(ctx context.Context) context.Context {
			middlewareCalled.Store(true)
			if !callbackCalled.Load() {
				middlewareBeforeCallback.Store(true)
			}
			return context.WithValue(ctx, "test-key", "test-value")
		}

		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnConnect: func(ctx context.Context, conn *websocket.Connection, params simbaModels.NoParams) error {
						callbackCalled.Store(true)
						// Check that middleware added value to context
						value := ctx.Value("test-key")
						assert.Equal(t, "test-value", value)
						return nil
					},
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return nil
					},
				}
			},
			websocket.WithMiddleware(testMiddleware),
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		time.Sleep(50 * time.Millisecond)

		assert.True(t, middlewareCalled.Load())
		assert.True(t, callbackCalled.Load())
		assert.True(t, middlewareBeforeCallback.Load())
	})

	t.Run("middleware runs for each message", func(t *testing.T) {
		var middlewareCount atomic.Int32

		testMiddleware := func(ctx context.Context) context.Context {
			middlewareCount.Add(1)
			return ctx
		}

		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return nil
					},
				}
			},
			websocket.WithMiddleware(testMiddleware),
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		// Send 3 messages
		for i := 0; i < 3; i++ {
			err = wsutil.WriteClientText(conn, []byte(fmt.Sprintf("message %d", i)))
			assert.NoError(t, err)
		}

		time.Sleep(100 * time.Millisecond)

		// Middleware should run once per message
		assert.Equal(t, int32(3), middlewareCount.Load())
	})

	t.Run("multiple middleware run in order", func(t *testing.T) {
		var order []string
		var mu sync.Mutex

		middleware1 := func(ctx context.Context) context.Context {
			mu.Lock()
			order = append(order, "mw1")
			mu.Unlock()
			return context.WithValue(ctx, "mw1", "value1")
		}

		middleware2 := func(ctx context.Context) context.Context {
			mu.Lock()
			order = append(order, "mw2")
			mu.Unlock()
			return context.WithValue(ctx, "mw2", "value2")
		}

		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						// Check both middleware values are in context
						assert.Equal(t, "value1", ctx.Value("mw1"))
						assert.Equal(t, "value2", ctx.Value("mw2"))
						return nil
					},
				}
			},
			websocket.WithMiddleware(middleware1, middleware2),
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		err = wsutil.WriteClientText(conn, []byte("test"))
		assert.NoError(t, err)

		time.Sleep(50 * time.Millisecond)

		mu.Lock()
		defer mu.Unlock()
		assert.Equal(t, 2, len(order))
		assert.Equal(t, "mw1", order[0])
		assert.Equal(t, "mw2", order[1])
	})

	t.Run("connectionID persists, traceID can change", func(t *testing.T) {
		var connIDs []string
		var traceIDs []string
		var mu sync.Mutex

		testMiddleware := func(ctx context.Context) context.Context {
			mu.Lock()
			defer mu.Unlock()

			if connID, ok := ctx.Value(simbaContext.ConnectionIDKey).(string); ok {
				connIDs = append(connIDs, connID)
			}

			// Add a test traceID
			ctx = context.WithValue(ctx, "traceID", fmt.Sprintf("trace-%d", len(traceIDs)))
			if traceID, ok := ctx.Value("traceID").(string); ok {
				traceIDs = append(traceIDs, traceID)
			}

			return ctx
		}

		handler := websocket.Handler(
			func() websocket.Callbacks[simbaModels.NoParams] {
				return websocket.Callbacks[simbaModels.NoParams]{
					OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
						return nil
					},
				}
			},
			websocket.WithMiddleware(testMiddleware),
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		// Send 3 messages
		for i := 0; i < 3; i++ {
			err = wsutil.WriteClientText(conn, []byte(fmt.Sprintf("message %d", i)))
			assert.NoError(t, err)
		}

		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		defer mu.Unlock()

		// All connectionIDs should be the same
		assert.Equal(t, 3, len(connIDs))
		for i := 1; i < len(connIDs); i++ {
			assert.Equal(t, connIDs[0], connIDs[i])
		}

		// All traceIDs should be different (new per message)
		assert.Equal(t, 3, len(traceIDs))
		assert.Equal(t, "trace-0", traceIDs[0])
		assert.Equal(t, "trace-1", traceIDs[1])
		assert.Equal(t, "trace-2", traceIDs[2])
	})
}
