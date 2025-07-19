package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sillen102/simba/middleware"
	"github.com/sillen102/simba/settings"
)

func TestCORS_AllowedOrigin(t *testing.T) {
	t.Parallel()

	cfg := settings.Cors{
		AllowedOrigins:   "http://example.com,http://another.com",
		AllowedMethods:   "GET,POST",
		AllowedHeaders:   "Content-Type,Authorization",
		AllowCredentials: true,
	}

	handler := middleware.CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
		t.Errorf("expected Access-Control-Allow-Origin to be 'http://example.com', got '%s'", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	t.Parallel()

	cfg := settings.Cors{
		AllowedOrigins:   "http://example.com",
		AllowedMethods:   "GET,POST",
		AllowedHeaders:   "Content-Type,Authorization",
		AllowCredentials: true,
	}

	handler := middleware.CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://notallowed.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("expected no Access-Control-Allow-Origin header, got '%s'", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_PreflightRequest(t *testing.T) {
	t.Parallel()

	cfg := settings.Cors{
		AllowedOrigins:   "http://example.com",
		AllowedMethods:   "GET,POST",
		AllowedHeaders:   "Content-Type,Authorization",
		AllowCredentials: true,
	}

	handler := middleware.CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status code %d, got %d", http.StatusNoContent, rec.Code)
	}

	if rec.Header().Get("Access-Control-Allow-Methods") != "GET,POST" {
		t.Errorf("expected Access-Control-Allow-Methods to be 'GET,POST', got '%s'", rec.Header().Get("Access-Control-Allow-Methods"))
	}
}

func TestCORS_NoOriginHeader(t *testing.T) {
	t.Parallel()

	cfg := settings.Cors{
		AllowedOrigins:   "http://example.com",
		AllowedMethods:   "GET,POST",
		AllowedHeaders:   "Content-Type,Authorization",
		AllowCredentials: true,
	}

	handler := middleware.CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("expected no Access-Control-Allow-Origin header, got '%s'", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}
