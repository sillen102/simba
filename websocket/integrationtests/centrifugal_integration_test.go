package integrationtests

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/centrifugal/centrifuge"
	"github.com/centrifugal/protocol"
	cws "github.com/coder/websocket"

	"github.com/sillen102/simba"
	ws "github.com/sillen102/simba/websocket"
)

type integrationAuthUser struct {
	ID string
}

func TestCentrifugalHandlerMountedInSimbaRouter_AllowsRPCOverWebsocket(t *testing.T) {
	t.Parallel()

	handler, err := ws.New(ws.Config{
		Websocket: centrifuge.WebsocketConfig{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		Setup: func(node *centrifuge.Node) error {
			node.OnConnecting(func(ctx context.Context, event centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
				return centrifuge.ConnectReply{
					Credentials: &centrifuge.Credentials{
						UserID: "integration-test-user",
					},
				}, nil
			})

			node.OnConnect(func(client *centrifuge.Client) {
				client.OnRPC(func(event centrifuge.RPCEvent, callback centrifuge.RPCCallback) {
					callback(centrifuge.RPCReply{
						Data: []byte(`{"echo":` + string(event.Data) + `}`),
					}, nil)
				})
			})

			return nil
		},
	})
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err = handler.Shutdown(shutdownCtx); err != nil {
			t.Fatalf("shutdown handler: %v", err)
		}
	})

	app := simba.New()
	app.Router.GET("/ws", handler)

	server := httptest.NewServer(app.Router)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := cws.Dial(ctx, toWebsocketURL(server.URL)+"/ws", nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer func() {
		_ = conn.Close(cws.StatusNormalClosure, "")
	}()

	connectReply := sendCommandAndReadReply(t, ctx, conn, &protocol.Command{
		Id:      1,
		Connect: &protocol.ConnectRequest{},
	})
	if connectReply.Error != nil {
		t.Fatalf("unexpected connect error: %+v", connectReply.Error)
	}
	if connectReply.Connect == nil {
		t.Fatal("expected connect result")
	}
	if connectReply.Connect.Client == "" {
		t.Fatal("expected server-assigned client id")
	}

	rpcPayload := []byte(`{"message":"hello simba"}`)
	rpcReply := sendCommandAndReadReply(t, ctx, conn, &protocol.Command{
		Id: 2,
		Rpc: &protocol.RPCRequest{
			Method: "echo",
			Data:   rpcPayload,
		},
	})
	if rpcReply.Error != nil {
		t.Fatalf("unexpected rpc error: %+v", rpcReply.Error)
	}
	if rpcReply.Rpc == nil {
		t.Fatal("expected rpc result")
	}
	if string(rpcReply.Rpc.Data) != `{"echo":{"message":"hello simba"}}` {
		t.Fatalf("unexpected rpc payload: %s", string(rpcReply.Rpc.Data))
	}
}

func TestAuthenticatedCentrifugalHandler_UsesSimbaAuthHandler(t *testing.T) {
	t.Parallel()

	bearerAuth := simba.BearerAuth(func(ctx context.Context, token string) (integrationAuthUser, error) {
		if token != "valid-token" {
			return integrationAuthUser{}, errors.New("invalid token")
		}
		return integrationAuthUser{ID: "user-123"}, nil
	}, simba.BearerAuthConfig{
		Name:        "BearerAuth",
		Format:      "JWT",
		Description: "Bearer token authentication",
	})

	handler, err := ws.NewAuthenticated(ws.AuthenticatedConfig[integrationAuthUser]{
		Config: ws.Config{
			Websocket: centrifuge.WebsocketConfig{
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			},
			Setup: func(node *centrifuge.Node) error {
				node.OnConnecting(func(ctx context.Context, event centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
					authUser, ok := ws.AuthModelFromContext[integrationAuthUser](ctx)
					if !ok {
						t.Fatal("expected auth user in connect context")
					}
					return centrifuge.ConnectReply{
						Credentials: &centrifuge.Credentials{UserID: authUser.ID},
					}, nil
				})

				node.OnConnect(func(client *centrifuge.Client) {
					client.OnRPC(func(event centrifuge.RPCEvent, callback centrifuge.RPCCallback) {
						authUser, ok := ws.AuthModelFromContext[integrationAuthUser](client.Context())
						if !ok {
							callback(centrifuge.RPCReply{}, errors.New("missing auth user"))
							return
						}
						payload, err := json.Marshal(map[string]string{
							"userId": authUser.ID,
						})
						if err != nil {
							callback(centrifuge.RPCReply{}, err)
							return
						}
						callback(centrifuge.RPCReply{Data: payload}, nil)
					})
				})
				return nil
			},
		},
		Auth: bearerAuth,
	})
	if err != nil {
		t.Fatalf("create authenticated handler: %v", err)
	}
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := handler.Shutdown(shutdownCtx); err != nil {
			t.Fatalf("shutdown handler: %v", err)
		}
	})

	app := simba.New()
	app.Router.GET("/ws-auth", handler)

	server := httptest.NewServer(app.Router)
	defer server.Close()

	unauthorizedResp, err := http.Get(server.URL + "/ws-auth")
	if err != nil {
		t.Fatalf("unauthorized request failed: %v", err)
	}
	defer unauthorizedResp.Body.Close()
	if unauthorizedResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing auth, got %d", unauthorizedResp.StatusCode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := cws.Dial(ctx, toWebsocketURL(server.URL)+"/ws-auth", &cws.DialOptions{
		HTTPHeader: http.Header{
			"Authorization": []string{"Bearer valid-token"},
		},
	})
	if err != nil {
		t.Fatalf("dial authenticated websocket: %v", err)
	}
	defer func() {
		_ = conn.Close(cws.StatusNormalClosure, "")
	}()

	connectReply := sendCommandAndReadReply(t, ctx, conn, &protocol.Command{
		Id:      1,
		Connect: &protocol.ConnectRequest{},
	})
	if connectReply.Error != nil {
		t.Fatalf("unexpected connect error: %+v", connectReply.Error)
	}

	rpcReply := sendCommandAndReadReply(t, ctx, conn, &protocol.Command{
		Id: 2,
		Rpc: &protocol.RPCRequest{
			Method: "whoami",
			Data:   []byte(`{}`),
		},
	})
	if rpcReply.Error != nil {
		t.Fatalf("unexpected rpc error: %+v", rpcReply.Error)
	}
	if string(rpcReply.Rpc.Data) != `{"userId":"user-123"}` {
		t.Fatalf("unexpected rpc payload: %s", string(rpcReply.Rpc.Data))
	}
}

func sendCommandAndReadReply(t *testing.T, ctx context.Context, conn *cws.Conn, cmd *protocol.Command) protocol.Reply {
	t.Helper()

	payload, err := json.Marshal(cmd)
	if err != nil {
		t.Fatalf("marshal command: %v", err)
	}

	if err = conn.Write(ctx, cws.MessageText, payload); err != nil {
		t.Fatalf("write command: %v", err)
	}

	typ, replyPayload, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read reply: %v", err)
	}
	if typ != cws.MessageText {
		t.Fatalf("expected text reply, got %v", typ)
	}

	var reply protocol.Reply
	if err = json.Unmarshal(replyPayload, &reply); err != nil {
		t.Fatalf("unmarshal reply %q: %v", string(replyPayload), err)
	}

	return reply
}

func toWebsocketURL(serverURL string) string {
	return "ws" + strings.TrimPrefix(serverURL, "http")
}
