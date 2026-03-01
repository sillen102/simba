package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/centrifugal/centrifuge"
	"github.com/sillen102/simba"
	"github.com/sillen102/simba/websocket"
)

const validToken = "valid-token"

type user struct {
	ID string
}

type rpcResponse struct {
	Method  string          `json:"method"`
	UserID  string          `json:"userId"`
	Payload json.RawMessage `json:"payload"`
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	app := simba.Default()

	bearerAuth := simba.BearerAuth(func(ctx context.Context, token string) (user, error) {
		if token != validToken {
			return user{}, centrifuge.ErrorUnauthorized
		}
		return user{ID: "authenticated-user"}, nil
	}, simba.BearerAuthConfig{
		Name:        "BearerAuth",
		Format:      "JWT",
		Description: "Bearer token authentication",
	})

	wsHandler, err := websocket.NewAuthenticated(websocket.AuthenticatedConfig[user]{
		Config: websocket.Config{
			Setup: func(node *centrifuge.Node) error {
				node.OnConnecting(func(ctx context.Context, event centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
					authUser, ok := websocket.AuthModelFromContext[user](ctx)
					if !ok {
						return centrifuge.ConnectReply{}, centrifuge.ErrorUnauthorized
					}

					slog.Info("client connecting", "client_id", event.ClientID, "user_id", authUser.ID)

					return centrifuge.ConnectReply{
						Credentials: &centrifuge.Credentials{UserID: authUser.ID},
					}, nil
				})

				node.OnConnect(func(client *centrifuge.Client) {
					slog.Info("client connected", "client_id", client.ID(), "user_id", client.UserID())

					client.OnRPC(func(event centrifuge.RPCEvent, callback centrifuge.RPCCallback) {
						payload, err := json.Marshal(rpcResponse{
							Method:  event.Method,
							UserID:  client.UserID(),
							Payload: json.RawMessage(event.Data),
						})
						if err != nil {
							callback(centrifuge.RPCReply{}, err)
							return
						}

						callback(centrifuge.RPCReply{Data: payload}, nil)
					})

					client.OnDisconnect(func(event centrifuge.DisconnectEvent) {
						slog.Info("client disconnected",
							"client_id", client.ID(),
							"user_id", client.UserID(),
							"code", event.Code,
							"reason", event.Reason)
					})
				})

				return nil
			},
		},
		Auth: bearerAuth,
	})
	if err != nil {
		panic(err)
	}

	app.Router.GET("/ws", wsHandler)
	app.RegisterShutdownHook(wsHandler.Shutdown)

	slog.Info("starting server", "addr", app.Server.Addr)
	slog.Info("websocket endpoint", "url", "ws://localhost:8080/ws")
	slog.Info("authorization header", "value", "Bearer valid-token")
	slog.Info("connect", "command", `{"id":1,"connect":{}}`)
	slog.Info("RPC echo", "command", `{"id":2,"rpc":{"method":"echo","data":{"message":"hello"}}}`)

	app.Start()
}
