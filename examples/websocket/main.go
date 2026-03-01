package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/centrifugal/centrifuge"
	"github.com/sillen102/simba"
	"github.com/sillen102/simba/websocket"
)

const validToken = "valid-token"

type rpcResponse struct {
	Method  string          `json:"method"`
	UserID  string          `json:"userId"`
	Payload json.RawMessage `json:"payload"`
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	app := simba.Default()

	wsHandler, err := websocket.New(websocket.Config{
		Websocket: centrifuge.WebsocketConfig{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		Setup: func(node *centrifuge.Node) error {
			node.OnConnecting(func(ctx context.Context, event centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
				userID := "anonymous"
				if event.Token != "" {
					if event.Token != validToken {
						return centrifuge.ConnectReply{}, centrifuge.ErrorUnauthorized
					}
					userID = "authenticated-user"
				}

				slog.Info("client connecting", "client_id", event.ClientID, "user_id", userID)

				return centrifuge.ConnectReply{
					Credentials: &centrifuge.Credentials{UserID: userID},
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
	})
	if err != nil {
		panic(err)
	}

	app.Router.GET("/ws", wsHandler)
	app.RegisterShutdownHook(wsHandler.Shutdown)

	slog.Info("starting server", "addr", app.Server.Addr)
	slog.Info("websocket endpoint", "url", "ws://localhost:8080/ws")
	slog.Info("connect anonymously", "command", `{"id":1,"connect":{}}`)
	slog.Info("connect with token", "command", `{"id":1,"connect":{"token":"valid-token"}}`)
	slog.Info("RPC echo", "command", `{"id":2,"rpc":{"method":"echo","data":{"message":"hello"}}}`)

	app.Start()
}
