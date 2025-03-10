package simba_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/simbaTest"
	"github.com/stretchr/testify/require"
)

func TestBasicAuth(t *testing.T) {
	t.Parallel()

	app := simba.Default()
	app.Router.POST("/test", simba.AuthJsonHandler(simbaTest.BasicAuthHandler, simbaTest.BasicAuthAuthenticationHandler))

	testCases := []struct {
		name           string
		username       string
		password       string
		expectedStatus int
	}{
		{
			name:           "valid credentials",
			username:       "user",
			password:       "password",
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "invalid credentials",
			username:       "user",
			password:       "invalid",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing credentials",
			username:       "",
			password:       "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			req.SetBasicAuth(tc.username, tc.password)

			w := httptest.NewRecorder()
			app.Router.ServeHTTP(w, req)

			resp := w.Result()
			require.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestApiKeyAuth(t *testing.T) {
	t.Parallel()

	app := simba.Default()
	app.Router.POST("/test", simba.AuthJsonHandler(simbaTest.ApiKeyAuthHandler, simbaTest.ApiKeyAuthAuthenticationHandler))

	testCases := []struct {
		name           string
		apiKey         string
		expectedStatus int
	}{
		{
			name:           "valid api key",
			apiKey:         "valid-key",
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "invalid api key",
			apiKey:         "invalid-key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing api key",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			req.Header.Add(simbaTest.ApiKeyAuthAuthenticationHandler.GetFieldName(), tc.apiKey)

			w := httptest.NewRecorder()
			app.Router.ServeHTTP(w, req)

			resp := w.Result()
			require.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestBearerTokenAuthHandler(t *testing.T) {
	t.Parallel()

	app := simba.Default()
	app.Router.POST("/test", simba.AuthJsonHandler(simbaTest.BearerTokenAuthHandler, simbaTest.BearerAuthAuthenticationHandler))

	testCases := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "valid token",
			token:          "Bearer token",
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "invalid token",
			token:          "Bearer invalid",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			req.Header.Set("Authorization", tc.token)

			w := httptest.NewRecorder()
			app.Router.ServeHTTP(w, req)

			resp := w.Result()
			require.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}
