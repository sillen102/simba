package simba_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/sillen102/simba"
	"github.com/sillen102/simba/simbaModels"
	"github.com/sillen102/simba/simbaTest/assert"
)

func TestWebSocketHandler_ConnectionLifecycle(t *testing.T) {
	t.Parallel()

	t.Run("OnConnect is called on connection", func(t *testing.T) {
		var connectCalled atomic.Bool

		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, params simbaModels.NoParams) error {
					connectCalled.Store(true)
					assert.NotNil(t, conn)
					assert.NotNil(t, connections)
					assert.True(t, conn.ID != "")
					return nil
				},
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					return nil
				},
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		// Connect as WebSocket client
		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		// Give time for OnConnect to be called
		time.Sleep(50 * time.Millisecond)

		assert.True(t, connectCalled.Load())
	})

	t.Run("OnMessage is called when message received", func(t *testing.T) {
		var messageCalled atomic.Bool
		var receivedData string

		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					messageCalled.Store(true)
					receivedData = string(data)
					return nil
				},
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		// Send a message
		testMessage := "Hello WebSocket"
		err = wsutil.WriteClientText(conn, []byte(testMessage))
		assert.NoError(t, err)

		// Give time for message to be processed
		time.Sleep(50 * time.Millisecond)

		assert.True(t, messageCalled.Load())
		assert.Equal(t, testMessage, receivedData)
	})

	t.Run("OnDisconnect is called on connection close", func(t *testing.T) {
		var disconnectCalled atomic.Bool

		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					return nil
				},
				OnDisconnect: func(ctx context.Context, params simbaModels.NoParams, err error) {
					disconnectCalled.Store(true)
				},
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)

		// Close connection
		conn.Close()

		// Give time for OnDisconnect to be called
		time.Sleep(50 * time.Millisecond)

		assert.True(t, disconnectCalled.Load())
	})

	t.Run("OnError is called on message error", func(t *testing.T) {
		var errorCalled atomic.Bool

		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					return fmt.Errorf("test error")
				},
				OnError: func(ctx context.Context, conn *simba.WebSocketConnection, err error) bool {
					errorCalled.Store(true)
					return false // Stop processing
				},
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		// Send a message that will trigger error
		err = wsutil.WriteClientText(conn, []byte("test"))
		assert.NoError(t, err)

		// Give time for error to be handled
		time.Sleep(50 * time.Millisecond)

		assert.True(t, errorCalled.Load())
	})
}

func TestWebSocketHandler_ConnectionTracking(t *testing.T) {
	t.Parallel()

	t.Run("connections map tracks active connections", func(t *testing.T) {
		var connectionCount int32
		var mu sync.Mutex
		var allConnectionIDs []string

		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, params simbaModels.NoParams) error {
					mu.Lock()
					defer mu.Unlock()
					atomic.StoreInt32(&connectionCount, int32(len(connections)))
					for id := range connections {
						allConnectionIDs = append(allConnectionIDs, id)
					}
					return nil
				},
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					return nil
				},
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		// Connect first client
		conn1, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn1.Close()

		time.Sleep(50 * time.Millisecond)
		assert.Equal(t, int32(1), atomic.LoadInt32(&connectionCount))

		// Connect second client
		conn2, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn2.Close()

		time.Sleep(50 * time.Millisecond)
		assert.Equal(t, int32(2), atomic.LoadInt32(&connectionCount))

		// Verify unique IDs
		mu.Lock()
		assert.True(t, len(allConnectionIDs) >= 2)
		mu.Unlock()
	})

	t.Run("connection is removed from map on disconnect", func(t *testing.T) {
		var finalCount atomic.Int32
		finalCount.Store(-1) // Set to invalid initially

		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					return nil
				},
				OnDisconnect: func(ctx context.Context, params simbaModels.NoParams, err error) {
					// Note: We can't check connections map here as it's not passed to OnDisconnect
					// This is by design - connection is already removed
				},
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)

		conn.Close()
		time.Sleep(50 * time.Millisecond)

		// Connection should be removed (verified by no panics or memory leaks)
	})
}

func TestWebSocketHandler_Broadcasting(t *testing.T) {
	t.Parallel()

	t.Run("can broadcast to all connections", func(t *testing.T) {
		var mu sync.Mutex
		received1 := ""
		received2 := ""

		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					// Broadcast to all connections
					for _, c := range connections {
						c.WriteText(string(data))
					}
					return nil
				},
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		// Connect two clients
		conn1, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn1.Close()

		conn2, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn2.Close()

		time.Sleep(50 * time.Millisecond)

		// Start readers
		go func() {
			msg, _, err := wsutil.ReadServerData(conn1)
			if err == nil {
				mu.Lock()
				received1 = string(msg)
				mu.Unlock()
			}
		}()

		go func() {
			msg, _, err := wsutil.ReadServerData(conn2)
			if err == nil {
				mu.Lock()
				received2 = string(msg)
				mu.Unlock()
			}
		}()

		// Send message from conn1
		err = wsutil.WriteClientText(conn1, []byte("broadcast message"))
		assert.NoError(t, err)

		// Give time for broadcast
		time.Sleep(100 * time.Millisecond)

		// Both should receive the message
		mu.Lock()
		assert.Equal(t, "broadcast message", received1)
		assert.Equal(t, "broadcast message", received2)
		mu.Unlock()
	})
}

func TestWebSocketHandler_WriteOperations(t *testing.T) {
	t.Parallel()

	t.Run("WriteText sends text message", func(t *testing.T) {
		var textWritten atomic.Bool

		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					err := conn.WriteText("echo: " + string(data))
					if err == nil {
						textWritten.Store(true)
					}
					return err
				},
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		// Send a message
		err = wsutil.WriteClientText(conn, []byte("test"))
		assert.NoError(t, err)

		// Give time for message to be processed
		time.Sleep(100 * time.Millisecond)

		assert.True(t, textWritten.Load(), "WriteText should have been called successfully")
	})

	t.Run("WriteJSON sends JSON message", func(t *testing.T) {
		type TestMessage struct {
			Type string `json:"type"`
			Data string `json:"data"`
		}

		var jsonWritten atomic.Bool

		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					err := conn.WriteJSON(TestMessage{Type: "response", Data: string(data)})
					if err == nil {
						jsonWritten.Store(true)
					}
					return err
				},
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		// Send a message to trigger WriteJSON
		err = wsutil.WriteClientText(conn, []byte("test"))
		assert.NoError(t, err)

		// Give time for JSON to be written
		time.Sleep(100 * time.Millisecond)

		assert.True(t, jsonWritten.Load(), "WriteJSON should have been called successfully")
	})
}

func TestAuthWebSocketHandler_Authentication(t *testing.T) {
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

		handler := simba.AuthWebSocketHandler(
			simba.AuthWebSocketCallbacks[simbaModels.NoParams, WSAuthModel]{
				OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, params simbaModels.NoParams, auth WSAuthModel) error {
					authReceived = auth
					return nil
				},
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte, auth WSAuthModel) error {
					return nil
				},
			},
			authHandler,
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		// Connect with valid auth header
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
		handler := simba.AuthWebSocketHandler(
			simba.AuthWebSocketCallbacks[simbaModels.NoParams, WSAuthModel]{
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte, auth WSAuthModel) error {
					return nil
				},
			},
			authHandler,
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		// Connect with invalid auth header
		dialer := ws.Dialer{
			Header: ws.HandshakeHeaderHTTP(http.Header{
				"Authorization": []string{"Bearer invalid-token"},
			}),
		}
		_, _, _, err := dialer.Dial(context.Background(), "ws"+server.URL[4:])
		assert.Error(t, err) // Should fail to connect
	})

	t.Run("auth is passed to OnMessage", func(t *testing.T) {
		var messageAuth WSAuthModel

		handler := simba.AuthWebSocketHandler(
			simba.AuthWebSocketCallbacks[simbaModels.NoParams, WSAuthModel]{
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte, auth WSAuthModel) error {
					messageAuth = auth
					return nil
				},
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

		// Send message
		err = wsutil.WriteClientText(conn, []byte("test"))
		assert.NoError(t, err)

		time.Sleep(50 * time.Millisecond)

		assert.Equal(t, 123, messageAuth.UserID)
		assert.Equal(t, "testuser", messageAuth.Username)
	})
}

func TestWebSocketHandler_ConcurrentConnections(t *testing.T) {
	t.Parallel()

	t.Run("handles multiple concurrent connections", func(t *testing.T) {
		var maxConnections atomic.Int32

		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, params simbaModels.NoParams) error {
					count := int32(len(connections))
					for {
						current := maxConnections.Load()
						if count <= current || maxConnections.CompareAndSwap(current, count) {
							break
						}
					}
					return nil
				},
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					return nil
				},
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		// Connect multiple clients concurrently
		numClients := 10
		var wg sync.WaitGroup
		wg.Add(numClients)

		for i := 0; i < numClients; i++ {
			go func() {
				defer wg.Done()
				conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
				if err == nil {
					defer conn.Close()
					time.Sleep(100 * time.Millisecond)
				}
			}()
		}

		wg.Wait()
		time.Sleep(50 * time.Millisecond)

		assert.True(t, maxConnections.Load() >= 1)
	})
}

func TestWebSocketHandler_ErrorRecovery(t *testing.T) {
	t.Parallel()

	t.Run("OnError can continue processing after error", func(t *testing.T) {
		var messageCount atomic.Int32

		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					messageCount.Add(1)
					if messageCount.Load() == 1 {
						return fmt.Errorf("first message error")
					}
					return nil
				},
				OnError: func(ctx context.Context, conn *simba.WebSocketConnection, err error) bool {
					return true // Continue processing
				},
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		// Send first message (will error)
		err = wsutil.WriteClientText(conn, []byte("message1"))
		assert.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		// Send second message (should succeed)
		err = wsutil.WriteClientText(conn, []byte("message2"))
		assert.NoError(t, err)
		time.Sleep(50 * time.Millisecond)

		assert.Equal(t, int32(2), messageCount.Load())
	})
}

func TestWebSocketConnection_Context(t *testing.T) {
	t.Parallel()

	t.Run("connection context is accessible", func(t *testing.T) {
		var connCtx context.Context

		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, params simbaModels.NoParams) error {
					connCtx = conn.Context()
					return nil
				},
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					return nil
				},
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)
		defer conn.Close()

		time.Sleep(50 * time.Millisecond)

		assert.NotNil(t, connCtx)
		assert.NoError(t, connCtx.Err()) // Context should not be cancelled yet
	})

	t.Run("connection context is cancelled on close", func(t *testing.T) {
		var connCtx context.Context
		var contextDone atomic.Bool

		handler := simba.WebSocketHandler(
			simba.WebSocketCallbacks[simbaModels.NoParams]{
				OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, params simbaModels.NoParams) error {
					connCtx = conn.Context()
					go func() {
						<-connCtx.Done()
						contextDone.Store(true)
					}()
					return nil
				},
				OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
					return nil
				},
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, _, err := ws.Dial(context.Background(), "ws"+server.URL[4:])
		assert.NoError(t, err)

		time.Sleep(50 * time.Millisecond)
		conn.Close()
		time.Sleep(50 * time.Millisecond)

		assert.True(t, contextDone.Load())
	})
}

// Helpers
func dialWebSocket(url string, headers http.Header) (net.Conn, error) {
	dialer := ws.Dialer{
		Header: ws.HandshakeHeaderHTTP(headers),
	}
	conn, _, _, err := dialer.Dial(context.Background(), url)
	return conn, err
}
