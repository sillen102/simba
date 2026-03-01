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

	"github.com/sillen102/simba/auth"
	"github.com/sillen102/simba/models"
	"github.com/sillen102/simba/simbaContext"
	"github.com/sillen102/simba/simbaTest/assert"
	simbawebsocket "github.com/sillen102/simba/websocket"

	"github.com/coder/websocket"
)

func TestHandler_ConnectionLifecycle(t *testing.T) {
	t.Parallel()

	t.Run("OnConnect is called on connection", func(t *testing.T) {
		t.Parallel()

		var connectCalled atomic.Bool
		var receivedConnID atomic.Value
		done := make(chan struct{})

		handler := simbawebsocket.Handler(
			func() simbawebsocket.Callbacks[models.NoParams] {
				return simbawebsocket.Callbacks[models.NoParams]{
					OnConnect: func(ctx context.Context, conn *simbawebsocket.Connection, params models.NoParams) error {
						connectCalled.Store(true)
						receivedConnID.Store(conn.ID)
						assert.NotNil(t, conn)
						assert.True(t, conn.ID != "")
						close(done)
						return nil
					},
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte) error {
						return nil
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)
		defer conn.CloseNow()

		<-done

		assert.True(t, connectCalled.Load())
		assert.True(t, receivedConnID.Load() != nil && receivedConnID.Load().(string) != "")
	})

	t.Run("OnMessage is called when message received", func(t *testing.T) {
		t.Parallel()

		var messageCalled atomic.Bool
		var receivedData atomic.Value
		done := make(chan struct{})

		handler := simbawebsocket.Handler(
			func() simbawebsocket.Callbacks[models.NoParams] {
				return simbawebsocket.Callbacks[models.NoParams]{
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte) error {
						messageCalled.Store(true)
						receivedData.Store(string(data))
						close(done)
						return nil
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)
		defer conn.CloseNow()

		testMessage := "Hello WebSocket"
		err = conn.Write(context.Background(), websocket.MessageText, []byte(testMessage))
		assert.NoError(t, err)

		<-done

		assert.True(t, messageCalled.Load())
		assert.Equal(t, testMessage, receivedData.Load().(string))
	})

	t.Run("OnMessage receives binary messages", func(t *testing.T) {
		t.Parallel()

		var receivedData atomic.Value
		done := make(chan struct{})

		handler := simbawebsocket.Handler(
			func() simbawebsocket.Callbacks[models.NoParams] {
				return simbawebsocket.Callbacks[models.NoParams]{
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte) error {
						receivedData.Store(data)
						close(done)
						return nil
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)
		defer conn.CloseNow()

		testData := []byte{0x01, 0x02, 0x03, 0x04}
		err = conn.Write(context.Background(), websocket.MessageBinary, testData)
		assert.NoError(t, err)

		<-done

		assert.Equal(t, testData, receivedData.Load().([]byte))
	})

	t.Run("OnDisconnect is called on connection close", func(t *testing.T) {
		t.Parallel()

		var disconnectCalled atomic.Bool
		var disconnectConnID atomic.Value
		done := make(chan struct{})

		handler := simbawebsocket.Handler(
			func() simbawebsocket.Callbacks[models.NoParams] {
				return simbawebsocket.Callbacks[models.NoParams]{
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte) error {
						return nil
					},
					OnDisconnect: func(ctx context.Context, connID string, params models.NoParams, err error) {
						disconnectCalled.Store(true)
						disconnectConnID.Store(connID)
						close(done)
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)

		conn.CloseNow()

		<-done

		assert.True(t, disconnectCalled.Load())
		assert.True(t, disconnectConnID.Load() != nil && disconnectConnID.Load().(string) != "")
	})

	t.Run("OnDisconnect receives connection ID", func(t *testing.T) {
		t.Parallel()

		var connectConnID atomic.Value
		var disconnectConnID atomic.Value
		connectDone := make(chan struct{})
		disconnectDone := make(chan struct{})

		handler := simbawebsocket.Handler(
			func() simbawebsocket.Callbacks[models.NoParams] {
				return simbawebsocket.Callbacks[models.NoParams]{
					OnConnect: func(ctx context.Context, conn *simbawebsocket.Connection, params models.NoParams) error {
						connectConnID.Store(conn.ID)
						close(connectDone)
						return nil
					},
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte) error {
						return nil
					},
					OnDisconnect: func(ctx context.Context, connID string, params models.NoParams, err error) {
						disconnectConnID.Store(connID)
						close(disconnectDone)
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)

		<-connectDone
		conn.CloseNow()
		<-disconnectDone

		assert.True(t, connectConnID.Load() != nil && connectConnID.Load().(string) != "")
		assert.Equal(t, connectConnID, disconnectConnID)
	})

	t.Run("OnError is called on message error", func(t *testing.T) {
		t.Parallel()

		var errorCalled atomic.Bool
		done := make(chan struct{})

		handler := simbawebsocket.Handler(
			func() simbawebsocket.Callbacks[models.NoParams] {
				return simbawebsocket.Callbacks[models.NoParams]{
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte) error {
						return fmt.Errorf("test error")
					},
					OnError: func(ctx context.Context, conn *simbawebsocket.Connection, err error) bool {
						errorCalled.Store(true)
						close(done)
						return false
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)
		defer conn.CloseNow()

		err = conn.Write(context.Background(), websocket.MessageText, []byte("test"))
		assert.NoError(t, err)

		<-done

		assert.True(t, errorCalled.Load())
	})
}

func TestHandler_ConnectionID(t *testing.T) {
	t.Parallel()

	t.Run("each connection gets unique ID", func(t *testing.T) {
		t.Parallel()

		var mu sync.Mutex
		var connIDs []string
		var wg sync.WaitGroup
		wg.Add(3)

		handler := simbawebsocket.Handler(
			func() simbawebsocket.Callbacks[models.NoParams] {
				return simbawebsocket.Callbacks[models.NoParams]{
					OnConnect: func(ctx context.Context, conn *simbawebsocket.Connection, params models.NoParams) error {
						mu.Lock()
						defer mu.Unlock()
						connIDs = append(connIDs, conn.ID)
						wg.Done()
						return nil
					},
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte) error {
						return nil
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn1, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)
		defer conn1.CloseNow()

		conn2, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)
		defer conn2.CloseNow()

		conn3, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)
		defer conn3.CloseNow()

		wg.Wait()

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
		t.Parallel()

		var connID atomic.Value
		done := make(chan struct{})

		handler := simbawebsocket.Handler(
			func() simbawebsocket.Callbacks[models.NoParams] {
				return simbawebsocket.Callbacks[models.NoParams]{
					OnConnect: func(ctx context.Context, conn *simbawebsocket.Connection, params models.NoParams) error {
						connID.Store(conn.ID)
						close(done)
						return nil
					},
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte) error {
						return nil
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)
		defer conn.CloseNow()

		<-done

		connIDStr := connID.Load().(string)
		assert.Equal(t, 36, len(connIDStr))
		assert.Equal(t, '-', rune(connIDStr[8]))
		assert.Equal(t, '-', rune(connIDStr[13]))
		assert.Equal(t, '-', rune(connIDStr[18]))
		assert.Equal(t, '-', rune(connIDStr[23]))
	})
}

func TestHandler_WriteOperations(t *testing.T) {
	t.Parallel()

	t.Run("WriteText sends text message", func(t *testing.T) {
		t.Parallel()

		// Echo back text messages
		handler := simbawebsocket.Handler(
			func() simbawebsocket.Callbacks[models.NoParams] {
				return simbawebsocket.Callbacks[models.NoParams]{
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte) error {
						return conn.WriteText(ctx, "echo: "+string(data))
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)
		defer conn.CloseNow()

		// Send a message to trigger the response
		err = conn.Write(context.Background(), websocket.MessageText, []byte("hello"))
		assert.NoError(t, err)

		readCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		msgType, msg, err := conn.Read(readCtx)
		assert.NoError(t, err)
		assert.Equal(t, websocket.MessageText, msgType)
		assert.Equal(t, "echo: hello", string(msg))
	})

	t.Run("WriteBinary sends binary message", func(t *testing.T) {
		t.Parallel()

		expectedData := []byte{0xDE, 0xAD, 0xBE, 0xEF}

		// Echo back binary data
		handler := simbawebsocket.Handler(
			func() simbawebsocket.Callbacks[models.NoParams] {
				return simbawebsocket.Callbacks[models.NoParams]{
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte) error {
						return conn.WriteBinary(ctx, expectedData)
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)
		defer conn.CloseNow()

		// Send a message to trigger the response
		err = conn.Write(context.Background(), websocket.MessageText, []byte("send binary"))
		assert.NoError(t, err)

		readCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		msgType, msg, err := conn.Read(readCtx)
		assert.NoError(t, err)
		assert.Equal(t, websocket.MessageBinary, msgType)
		assert.Equal(t, expectedData, msg)
	})

	t.Run("WriteJSON sends JSON message", func(t *testing.T) {
		t.Parallel()

		type TestMessage struct {
			Type string `json:"type"`
			Data string `json:"data"`
		}

		handler := simbawebsocket.Handler(
			func() simbawebsocket.Callbacks[models.NoParams] {
				return simbawebsocket.Callbacks[models.NoParams]{
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte) error {
						return conn.WriteJSON(ctx, TestMessage{Type: "response", Data: string(data)})
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)
		defer conn.CloseNow()

		// Send a message to trigger the response
		err = conn.Write(context.Background(), websocket.MessageText, []byte("hello"))
		assert.NoError(t, err)

		readCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		msgType, msg, err := conn.Read(readCtx)
		assert.NoError(t, err)
		assert.Equal(t, websocket.MessageText, msgType)
		assert.Equal(t, `{"type":"response","data":"hello"}`, string(msg))
	})
}

func TestHandler_ExternalRegistry(t *testing.T) {
	t.Parallel()

	t.Run("can implement external connection registry", func(t *testing.T) {
		t.Parallel()

		registry := &sync.Map{}
		connectDone := make(chan struct{})
		disconnectDone := make(chan struct{})

		handler := simbawebsocket.Handler(
			func() simbawebsocket.Callbacks[models.NoParams] {
				return simbawebsocket.Callbacks[models.NoParams]{
					OnConnect: func(ctx context.Context, conn *simbawebsocket.Connection, params models.NoParams) error {
						registry.Store(conn.ID, conn)
						close(connectDone)
						return nil
					},
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte) error {
						return nil
					},
					OnDisconnect: func(ctx context.Context, connID string, params models.NoParams, err error) {
						registry.Delete(connID)
						close(disconnectDone)
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)

		<-connectDone

		count := 0
		registry.Range(func(key, value any) bool {
			count++
			return true
		})
		assert.Equal(t, 1, count)

		conn.CloseNow()
		<-disconnectDone

		count = 0
		registry.Range(func(key, value any) bool {
			count++
			return true
		})
		assert.Equal(t, 0, count)
	})

	t.Run("can send to connection from external source", func(t *testing.T) {
		t.Parallel()

		registry := &sync.Map{}
		var registeredConnID atomic.Value
		done := make(chan struct{})

		handler := simbawebsocket.Handler(
			func() simbawebsocket.Callbacks[models.NoParams] {
				return simbawebsocket.Callbacks[models.NoParams]{
					OnConnect: func(ctx context.Context, conn *simbawebsocket.Connection, params models.NoParams) error {
						registeredConnID.Store(conn.ID)
						registry.Store(conn.ID, conn)
						close(done)
						return nil
					},
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte) error {
						return nil
					},
					OnDisconnect: func(ctx context.Context, connID string, params models.NoParams, err error) {
						registry.Delete(connID)
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)
		defer conn.CloseNow()

		<-done

		readCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if wsConn, ok := registry.Load(registeredConnID.Load().(string)); ok {
			wsConn.(*simbawebsocket.Connection).WriteText(context.Background(), "external message")
		}

		_, msg, err := conn.Read(readCtx)
		assert.NoError(t, err)
		assert.Equal(t, "external message", string(msg))
	})
}

func TestAuthHandler_Authentication(t *testing.T) {
	t.Parallel()

	authHandler := auth.BearerAuth(
		func(ctx context.Context, token string) (WSAuthModel, error) {
			if token == "valid-token" {
				return WSAuthModel{
					UserID:   123,
					Username: "testuser",
				}, nil
			}
			return WSAuthModel{}, fmt.Errorf("invalid token")
		},
		auth.BearerAuthConfig{
			Name:        "BearerAuth",
			Format:      "JWT",
			Description: "Test bearer auth",
		},
	)

	t.Run("authenticated connection succeeds with valid token", func(t *testing.T) {
		t.Parallel()

		var authReceived atomic.Value
		done := make(chan struct{})

		handler := simbawebsocket.AuthHandler(
			func() simbawebsocket.AuthCallbacks[models.NoParams, WSAuthModel] {
				return simbawebsocket.AuthCallbacks[models.NoParams, WSAuthModel]{
					OnConnect: func(ctx context.Context, conn *simbawebsocket.Connection, params models.NoParams, auth WSAuthModel) error {
						authReceived.Store(auth)
						close(done)
						return nil
					},
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte, auth WSAuthModel) error {
						return nil
					},
				}
			},
			authHandler,
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], &websocket.DialOptions{
			HTTPHeader: http.Header{
				"Authorization": {"Bearer valid-token"},
			},
		})
		assert.NoError(t, err)
		defer conn.CloseNow()

		<-done

		assert.Equal(t, 123, authReceived.Load().(WSAuthModel).UserID)
		assert.Equal(t, "testuser", authReceived.Load().(WSAuthModel).Username)
	})

	t.Run("connection rejected with invalid token", func(t *testing.T) {
		t.Parallel()

		handler := simbawebsocket.AuthHandler(
			func() simbawebsocket.AuthCallbacks[models.NoParams, WSAuthModel] {
				return simbawebsocket.AuthCallbacks[models.NoParams, WSAuthModel]{
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte, auth WSAuthModel) error {
						return nil
					},
				}
			},
			authHandler,
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		_, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], &websocket.DialOptions{
			HTTPHeader: http.Header{
				"Authorization": {"Bearer invalid-token"},
			},
		})
		assert.Error(t, err)
	})

	t.Run("auth is passed to OnMessage", func(t *testing.T) {
		t.Parallel()

		var messageAuth atomic.Value
		done := make(chan struct{})

		handler := simbawebsocket.AuthHandler(
			func() simbawebsocket.AuthCallbacks[models.NoParams, WSAuthModel] {
				return simbawebsocket.AuthCallbacks[models.NoParams, WSAuthModel]{
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte, auth WSAuthModel) error {
						messageAuth.Store(auth)
						close(done)
						return nil
					},
				}
			},
			authHandler,
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], &websocket.DialOptions{
			HTTPHeader: http.Header{
				"Authorization": {"Bearer valid-token"},
			},
		})
		assert.NoError(t, err)
		defer conn.CloseNow()

		err = conn.Write(context.Background(), websocket.MessageText, []byte("test"))
		assert.NoError(t, err)

		<-done

		assert.Equal(t, 123, messageAuth.Load().(WSAuthModel).UserID)
		assert.Equal(t, "testuser", messageAuth.Load().(WSAuthModel).Username)
	})

	t.Run("auth is passed to OnDisconnect", func(t *testing.T) {
		t.Parallel()

		var disconnectAuth atomic.Value
		var disconnectConnID atomic.Value
		connectDone := make(chan struct{})
		disconnectDone := make(chan struct{})

		handler := simbawebsocket.AuthHandler(
			func() simbawebsocket.AuthCallbacks[models.NoParams, WSAuthModel] {
				return simbawebsocket.AuthCallbacks[models.NoParams, WSAuthModel]{
					OnConnect: func(ctx context.Context, conn *simbawebsocket.Connection, params models.NoParams, auth WSAuthModel) error {
						close(connectDone)
						return nil
					},
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte, auth WSAuthModel) error {
						return nil
					},
					OnDisconnect: func(ctx context.Context, connID string, params models.NoParams, auth WSAuthModel, err error) {
						disconnectConnID.Store(connID)
						disconnectAuth.Store(auth)
						close(disconnectDone)
					},
				}
			},
			authHandler,
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], &websocket.DialOptions{
			HTTPHeader: http.Header{
				"Authorization": {"Bearer valid-token"},
			},
		})
		assert.NoError(t, err)

		<-connectDone
		conn.CloseNow()
		<-disconnectDone

		assert.True(t, disconnectConnID.Load() != nil && disconnectConnID.Load().(string) != "")
		assert.Equal(t, 123, disconnectAuth.Load().(WSAuthModel).UserID)
		assert.Equal(t, "testuser", disconnectAuth.Load().(WSAuthModel).Username)
	})
}

func TestHandler_ConcurrentConnections(t *testing.T) {
	t.Parallel()

	t.Run("handles multiple concurrent connections", func(t *testing.T) {
		t.Parallel()

		var connectionCount atomic.Int32
		var connectWg sync.WaitGroup
		var disconnectWg sync.WaitGroup
		numClients := 10
		connectWg.Add(numClients)
		disconnectWg.Add(numClients)

		handler := simbawebsocket.Handler(
			func() simbawebsocket.Callbacks[models.NoParams] {
				return simbawebsocket.Callbacks[models.NoParams]{
					OnConnect: func(ctx context.Context, conn *simbawebsocket.Connection, params models.NoParams) error {
						connectionCount.Add(1)
						connectWg.Done()
						return nil
					},
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte) error {
						return nil
					},
					OnDisconnect: func(ctx context.Context, connID string, params models.NoParams, err error) {
						connectionCount.Add(-1)
						disconnectWg.Done()
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		var wg sync.WaitGroup
		wg.Add(numClients)

		conns := make([]*websocket.Conn, numClients)
		var connsMu sync.Mutex

		for i := 0; i < numClients; i++ {
			go func(idx int) {
				defer wg.Done()
				conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
				if err == nil {
					connsMu.Lock()
					conns[idx] = conn
					connsMu.Unlock()
				}
			}(i)
		}

		wg.Wait()
		connectWg.Wait()

		assert.Equal(t, int32(numClients), connectionCount.Load())

		for _, conn := range conns {
			if conn != nil {
				conn.CloseNow()
			}
		}

		disconnectWg.Wait()
		assert.Equal(t, int32(0), connectionCount.Load())
	})
}

func TestHandler_ErrorRecovery(t *testing.T) {
	t.Parallel()

	t.Run("OnError can continue processing after error", func(t *testing.T) {
		t.Parallel()

		var messageCount atomic.Int32
		var wg sync.WaitGroup
		wg.Add(2)

		handler := simbawebsocket.Handler(
			func() simbawebsocket.Callbacks[models.NoParams] {
				return simbawebsocket.Callbacks[models.NoParams]{
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte) error {
						messageCount.Add(1)
						wg.Done()
						if messageCount.Load() == 1 {
							return fmt.Errorf("first message error")
						}
						return nil
					},
					OnError: func(ctx context.Context, conn *simbawebsocket.Connection, err error) bool {
						return true
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)
		defer conn.CloseNow()

		err = conn.Write(context.Background(), websocket.MessageText, []byte("message1"))
		assert.NoError(t, err)

		err = conn.Write(context.Background(), websocket.MessageText, []byte("message2"))
		assert.NoError(t, err)

		wg.Wait()

		assert.Equal(t, int32(2), messageCount.Load())
	})

	t.Run("connection closes when OnError returns false", func(t *testing.T) {
		t.Parallel()

		var disconnectCalled atomic.Bool
		done := make(chan struct{})

		handler := simbawebsocket.Handler(
			func() simbawebsocket.Callbacks[models.NoParams] {
				return simbawebsocket.Callbacks[models.NoParams]{
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte) error {
						return fmt.Errorf("error")
					},
					OnError: func(ctx context.Context, conn *simbawebsocket.Connection, err error) bool {
						return false
					},
					OnDisconnect: func(ctx context.Context, connID string, params models.NoParams, err error) {
						disconnectCalled.Store(true)
						close(done)
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)
		defer conn.CloseNow()

		err = conn.Write(context.Background(), websocket.MessageText, []byte("test"))
		assert.NoError(t, err)

		<-done

		assert.True(t, disconnectCalled.Load())
	})
}

func TestHandler_ThreadSafety(t *testing.T) {
	t.Parallel()

	t.Run("concurrent writes are thread-safe", func(t *testing.T) {
		t.Parallel()

		var wsConn *simbawebsocket.Connection
		var connReady sync.WaitGroup
		connReady.Add(1)

		handler := simbawebsocket.Handler(
			func() simbawebsocket.Callbacks[models.NoParams] {
				return simbawebsocket.Callbacks[models.NoParams]{
					OnConnect: func(ctx context.Context, conn *simbawebsocket.Connection, params models.NoParams) error {
						wsConn = conn
						connReady.Done()
						return nil
					},
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte) error {
						return nil
					},
				}
			},
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)
		defer conn.CloseNow()

		connReady.Wait()

		var wg sync.WaitGroup
		numWriters := 10
		messagesPerWriter := 100

		for i := 0; i < numWriters; i++ {
			wg.Add(1)
			go func(writerID int) {
				defer wg.Done()
				for j := 0; j < messagesPerWriter; j++ {
					wsConn.WriteText(context.Background(), fmt.Sprintf("writer-%d-msg-%d", writerID, j))
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
		t.Parallel()

		var middlewareCalled atomic.Bool
		var callbackCalled atomic.Bool
		var middlewareBeforeCallback atomic.Bool
		done := make(chan struct{})

		testMiddleware := func(ctx context.Context) context.Context {
			middlewareCalled.Store(true)
			if !callbackCalled.Load() {
				middlewareBeforeCallback.Store(true)
			}
			return context.WithValue(ctx, "test-key", "test-value")
		}

		handler := simbawebsocket.Handler(
			func() simbawebsocket.Callbacks[models.NoParams] {
				return simbawebsocket.Callbacks[models.NoParams]{
					OnConnect: func(ctx context.Context, conn *simbawebsocket.Connection, params models.NoParams) error {
						callbackCalled.Store(true)
						// Check that middleware added value to context
						value := ctx.Value("test-key")
						assert.Equal(t, "test-value", value)
						close(done)
						return nil
					},
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte) error {
						return nil
					},
				}
			},
			simbawebsocket.WithMiddleware(testMiddleware),
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)
		defer conn.CloseNow()

		<-done

		assert.True(t, middlewareCalled.Load())
		assert.True(t, callbackCalled.Load())
		assert.True(t, middlewareBeforeCallback.Load())
	})

	t.Run("middleware runs for each message", func(t *testing.T) {
		t.Parallel()

		var middlewareCount atomic.Int32
		var wg sync.WaitGroup
		wg.Add(3)

		testMiddleware := func(ctx context.Context) context.Context {
			middlewareCount.Add(1)
			return ctx
		}

		handler := simbawebsocket.Handler(
			func() simbawebsocket.Callbacks[models.NoParams] {
				return simbawebsocket.Callbacks[models.NoParams]{
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte) error {
						wg.Done()
						return nil
					},
				}
			},
			simbawebsocket.WithMiddleware(testMiddleware),
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)
		defer conn.CloseNow()

		// Send 3 messages
		for i := 0; i < 3; i++ {
			err = conn.Write(context.Background(), websocket.MessageText, []byte(fmt.Sprintf("message %d", i)))
			assert.NoError(t, err)
		}

		wg.Wait()

		// Middleware should run once per message
		assert.Equal(t, int32(3), middlewareCount.Load())
	})

	t.Run("multiple middleware run in order", func(t *testing.T) {
		t.Parallel()

		var order []string
		var mu sync.Mutex
		done := make(chan struct{})

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

		handler := simbawebsocket.Handler(
			func() simbawebsocket.Callbacks[models.NoParams] {
				return simbawebsocket.Callbacks[models.NoParams]{
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte) error {
						// Check both middleware values are in context
						assert.Equal(t, "value1", ctx.Value("mw1"))
						assert.Equal(t, "value2", ctx.Value("mw2"))
						close(done)
						return nil
					},
				}
			},
			simbawebsocket.WithMiddleware(middleware1, middleware2),
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)
		defer conn.CloseNow()

		err = conn.Write(context.Background(), websocket.MessageText, []byte("test"))
		assert.NoError(t, err)

		<-done

		mu.Lock()
		defer mu.Unlock()
		assert.Equal(t, 2, len(order))
		assert.Equal(t, "mw1", order[0])
		assert.Equal(t, "mw2", order[1])
	})

	t.Run("connectionID persists, traceID can change", func(t *testing.T) {
		t.Parallel()

		var connIDs []string
		var traceIDs []string
		var mu sync.Mutex
		var wg sync.WaitGroup
		wg.Add(3)

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

		handler := simbawebsocket.Handler(
			func() simbawebsocket.Callbacks[models.NoParams] {
				return simbawebsocket.Callbacks[models.NoParams]{
					OnMessage: func(ctx context.Context, conn *simbawebsocket.Connection, data []byte) error {
						wg.Done()
						return nil
					},
				}
			},
			simbawebsocket.WithMiddleware(testMiddleware),
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		conn, _, err := websocket.Dial(context.Background(), "ws"+server.URL[4:], nil)
		assert.NoError(t, err)
		defer conn.CloseNow()

		// Send 3 messages
		for i := 0; i < 3; i++ {
			err = conn.Write(context.Background(), websocket.MessageText, []byte(fmt.Sprintf("message %d", i)))
			assert.NoError(t, err)
		}

		wg.Wait()

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
