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
	Route   string          `json:"route"`
	Method  string          `json:"method"`
	UserID  string          `json:"userId"`
	Payload json.RawMessage `json:"payload"`
}

var authHandler = simba.BearerAuth(authenticateUser, simba.BearerAuthConfig{
	Name:        "BearerAuth",
	Format:      "JWT",
	Description: "Bearer token authentication",
})

func main() {
	configureLogging()

	app := simba.Default()

	publicHandler := newPublicWebsocketHandler()
	privateRouteHandler := websocket.AuthHandler(newAuthenticatedWebsocketHandler, authHandler)

	app.Router.GET("/ws/public", publicHandler)
	app.Router.GET("/ws/private", privateRouteHandler)
	app.RegisterShutdownHook(publicHandler.Shutdown)
	app.RegisterShutdownHook(privateRouteHandler.Shutdown)

	slog.Info("starting server", "addr", app.Server.Addr)
	slog.Info("public websocket endpoint", "url", "ws://localhost:8080/ws/public")
	slog.Info("private websocket endpoint", "url", "ws://localhost:8080/ws/private")
	slog.Info("private auth header", "value", "Bearer valid-token")
	slog.Info("connect command", "command", `{"id":1,"connect":{}}`)
	slog.Info("RPC command", "command", `{"id":2,"rpc":{"method":"echo","data":{"message":"hello"}}}`)
	slog.Info("server publish example", "code", `_, _ = privateRouteHandler.Publish("user:authenticated-user", []byte(`+"`"+`{"type":"notification"}`+"`"+`))`)

	app.Start()
}

func configureLogging() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)
}

func newPublicWebsocketHandler() *websocket.Handler {
	return (websocket.Config{
		OnConnecting: websocket.OnConnecting(func(ctx context.Context, event centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
			slog.Info("public client connecting", "client_id", event.ClientID)

			return centrifuge.ConnectReply{
				Credentials: &centrifuge.Credentials{UserID: "anonymous"},
			}, nil
		}),
		OnConnect: websocket.OnConnect(func(client *centrifuge.Client) {
			slog.Info("public client connected", "client_id", client.ID(), "user_id", client.UserID())

			client.OnRPC(func(event centrifuge.RPCEvent, callback centrifuge.RPCCallback) {
				payload, err := json.Marshal(rpcResponse{
					Route:   "public",
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
		}),
	}).Handler()
}

func newAuthenticatedWebsocketHandler() *websocket.AuthenticatedHandler[user] {
	return (websocket.AuthenticatedConfig[user]{
		OnConnecting: websocket.OnConnecting(func(ctx context.Context, event centrifuge.ConnectEvent, authUser *user) (centrifuge.ConnectReply, error) {
			if authUser == nil {
				return centrifuge.ConnectReply{}, centrifuge.ErrorUnauthorized
			}

			slog.Info("private client connecting", "client_id", event.ClientID, "user_id", authUser.ID)

			return centrifuge.ConnectReply{
				Credentials: &centrifuge.Credentials{UserID: authUser.ID},
				Subscriptions: map[string]centrifuge.SubscribeOptions{
					userChannel(authUser.ID): {},
				},
			}, nil
		}),
		OnConnect: websocket.OnConnect(func(client *centrifuge.Client) {
			slog.Info("private client connected", "client_id", client.ID(), "user_id", client.UserID())

			client.OnRPC(func(event centrifuge.RPCEvent, callback centrifuge.RPCCallback) {
				payload, err := json.Marshal(rpcResponse{
					Route:   "private",
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
				slog.Info("private client disconnected",
					"client_id", client.ID(),
					"user_id", client.UserID(),
					"code", event.Code,
					"reason", event.Reason)
			})
		}),
	}).Handler()
}

func authenticateUser(ctx context.Context, token string) (user, error) {
	if token != validToken {
		return user{}, centrifuge.ErrorUnauthorized
	}
	return user{ID: "authenticated-user"}, nil
}

func userChannel(userID string) string {
	return "user:" + userID
}
