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

	var setupCalled bool

	handler, err := websocket.New(websocket.Config{
		Setup: func(node *centrifuge.Node) error {
			setupCalled = true
			node.OnConnecting(func(ctx context.Context, event centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
				return centrifuge.ConnectReply{}, nil
			})
			return nil
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !setupCalled {
		t.Fatal("expected setup callback to run")
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

func TestNewAuthenticated(t *testing.T) {
	t.Parallel()

	handler, err := websocket.NewAuthenticated(websocket.AuthenticatedConfig[testAuthModel]{
		Config: websocket.Config{
			Setup: func(node *centrifuge.Node) error { return nil },
		},
		Auth: testAuthHandler{model: testAuthModel{Name: "test-user"}},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got, ok := handler.GetAuthModel().(testAuthModel); !ok || got != (testAuthModel{}) {
		t.Fatalf("expected auth model metadata, got %#v", handler.GetAuthModel())
	}
	if handler.GetAuthHandler() == nil {
		t.Fatal("expected auth handler metadata")
	}

	if err := handler.Shutdown(context.Background()); err != nil {
		t.Fatalf("expected clean shutdown, got %v", err)
	}
}

func TestNewAuthenticatedRequiresAuthHandler(t *testing.T) {
	t.Parallel()

	handler, err := websocket.NewAuthenticated[testAuthModel](websocket.AuthenticatedConfig[testAuthModel]{})
	if err == nil {
		t.Fatal("expected error")
	}
	if handler != nil {
		t.Fatal("expected nil handler")
	}
}
