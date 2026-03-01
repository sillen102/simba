package websocket_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/centrifugal/centrifuge"

	"github.com/sillen102/simba/websocket"
)

func TestNew(t *testing.T) {
	t.Parallel()

	handler, err := websocket.New(websocket.Config{
		OnConnecting: websocket.OnConnecting(func(ctx context.Context, event centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
			return centrifuge.ConnectReply{}, nil
		}),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if handler == nil {
		t.Fatal("expected handler")
	}
	if handler.Node() == nil {
		t.Fatal("expected node")
	}
	if handler.HTTPHandler() == nil {
		t.Fatal("expected http handler")
	}

	if err := handler.Shutdown(context.Background()); err != nil {
		t.Fatalf("expected clean shutdown, got %v", err)
	}
}

func TestConfigHandler(t *testing.T) {
	t.Parallel()

	handler := (websocket.Config{
		OnConnecting: websocket.OnConnecting(func(ctx context.Context, event centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
			return centrifuge.ConnectReply{}, nil
		}),
	}).Handler()

	if handler == nil {
		t.Fatal("expected handler")
	}

	if err := handler.Shutdown(context.Background()); err != nil {
		t.Fatalf("expected clean shutdown, got %v", err)
	}
}

func TestNewSetupError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("boom")

	handler, err := websocket.New(websocket.Config{
		Setup: func(node *centrifuge.Node) error {
			return expectedErr
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected wrapped setup error, got %v", err)
	}
	if handler != nil {
		t.Fatal("expected nil handler on setup error")
	}
}

type testAuthHandler struct {
	model testAuthModel
	err   error
}

type testAuthModel struct {
	Name string
}

func (h testAuthHandler) GetHandler() func(r *http.Request) (testAuthModel, error) {
	return func(r *http.Request) (testAuthModel, error) {
		return h.model, h.err
	}
}

func TestNewAuthenticatedHandler(t *testing.T) {
	t.Parallel()

	handler, err := websocket.NewAuthenticatedHandler(websocket.AuthenticatedConfig[testAuthModel]{
		OnConnecting: websocket.OnConnecting(func(ctx context.Context, event centrifuge.ConnectEvent, authModel *testAuthModel) (centrifuge.ConnectReply, error) {
			if authModel == nil || authModel.Name != "test-user" {
				t.Fatalf("expected injected auth model, got %#v", authModel)
			}
			return centrifuge.ConnectReply{}, nil
		}),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if handler.GetAuthHandler() != nil {
		t.Fatal("expected no auth handler before wrapping")
	}
	if got, ok := handler.GetAuthModel().(testAuthModel); !ok || got != (testAuthModel{}) {
		t.Fatalf("expected auth model metadata, got %#v", handler.GetAuthModel())
	}

	if err := handler.Shutdown(context.Background()); err != nil {
		t.Fatalf("expected clean shutdown, got %v", err)
	}
}

func TestAuthenticatedConfigHandler(t *testing.T) {
	t.Parallel()

	handler := (websocket.AuthenticatedConfig[testAuthModel]{}).Handler()

	if handler == nil {
		t.Fatal("expected handler")
	}

	if err := handler.Shutdown(context.Background()); err != nil {
		t.Fatalf("expected clean shutdown, got %v", err)
	}
}

func TestAuthHandler(t *testing.T) {
	t.Parallel()

	handler, err := websocket.NewAuthenticatedHandler(websocket.AuthenticatedConfig[testAuthModel]{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	wrapped := websocket.NewAuthenticated(handler, testAuthHandler{model: testAuthModel{Name: "test-user"}})

	if got, ok := handler.GetAuthModel().(testAuthModel); !ok || got != (testAuthModel{}) {
		t.Fatalf("expected original handler metadata, got %#v", handler.GetAuthModel())
	}
	if handler.GetAuthHandler() != nil {
		t.Fatal("expected original handler auth metadata to remain unset")
	}

	if got, ok := wrapped.GetAuthModel().(testAuthModel); !ok || got != (testAuthModel{}) {
		t.Fatalf("expected wrapped auth model metadata, got %#v", wrapped.GetAuthModel())
	}
	if wrapped.GetAuthHandler() == nil {
		t.Fatal("expected auth handler metadata")
	}

	if err := wrapped.Shutdown(context.Background()); err != nil {
		t.Fatalf("expected clean shutdown, got %v", err)
	}
}

func TestAuthHandlerFactory(t *testing.T) {
	t.Parallel()

	wrapped := websocket.AuthHandler(func() *websocket.AuthenticatedHandler[testAuthModel] {
		return (websocket.AuthenticatedConfig[testAuthModel]{}).Handler()
	}, testAuthHandler{model: testAuthModel{Name: "test-user"}})

	if got, ok := wrapped.GetAuthModel().(testAuthModel); !ok || got != (testAuthModel{}) {
		t.Fatalf("expected wrapped auth model metadata, got %#v", wrapped.GetAuthModel())
	}
	if wrapped.GetAuthHandler() == nil {
		t.Fatal("expected auth handler metadata")
	}

	if err := wrapped.Shutdown(context.Background()); err != nil {
		t.Fatalf("expected clean shutdown, got %v", err)
	}
}
