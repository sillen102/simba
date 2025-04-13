package settings_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/simbaModels"
	"github.com/sillen102/simba/simbaTestAssert"
)

func TestLoadApplicationNameDefault(t *testing.T) {
	t.Parallel()
	s, err := settings.Load()
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, "Simba Application", s.Name)
}

func TestLoadApplicationNameFromEnvironment(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithEnvGetter(mockEnvGetter("APPLICATION_NAME", "Mock App")))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, "Mock App", s.Name)
}

func TestWithApplicationName(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithApplicationName("Custom App"))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, "Custom App", s.Name)
}

func TestLoadApplicationVersionDefault(t *testing.T) {
	t.Parallel()
	s, err := settings.Load()
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, "0.1.0", s.Version)
}

func TestLoadApplicationVersionFromEnvironment(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithEnvGetter(mockEnvGetter("APPLICATION_VERSION", "2.0.0")))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, "2.0.0", s.Version)
}

func TestWithApplicationVersion(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithApplicationVersion("1.0.0"))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, "1.0.0", s.Version)
}

func TestLoadServerHostDefault(t *testing.T) {
	t.Parallel()
	s, err := settings.Load()
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, "0.0.0.0", s.Host)
}

func TestLoadServerHostFromEnvironment(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithEnvGetter(mockEnvGetter("SIMBA_SERVER_HOST", "127.0.0.2")))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, "127.0.0.2", s.Host)
}

func TestWithServerHost(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithServerHost("127.0.0.3"))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, "127.0.0.3", s.Host)
}

func TestLoadServerPortDefault(t *testing.T) {
	t.Parallel()
	s, err := settings.Load()
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, 9999, s.Port)
}

func TestLoadServerPortFromEnvironment(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithEnvGetter(mockEnvGetter("SIMBA_SERVER_PORT", "9000")))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, 9000, s.Port)
}

func TestWithServerPort(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithServerPort(8080))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, 8080, s.Port)
}

func TestLoadAllowUnknownFieldsDefault(t *testing.T) {
	t.Parallel()
	s, err := settings.Load()
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.True(t, s.AllowUnknownFields)
}

func TestLoadAllowUnknownFieldsFromEnvironment(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithEnvGetter(mockEnvGetter("SIMBA_REQUEST_ALLOW_UNKNOWN_FIELDS", "false")))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.False(t, s.AllowUnknownFields)
}

func TestWithAllowUnknownFields(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithAllowUnknownFields(false))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.False(t, s.AllowUnknownFields)
}

func TestLoadLogRequestBodyDefault(t *testing.T) {
	t.Parallel()
	s, err := settings.Load()
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.False(t, s.LogRequestBody)
}

func TestLoadLogRequestBodyFromEnvironment(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithEnvGetter(mockEnvGetter("SIMBA_REQUEST_LOG_REQUEST_BODY", "true")))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.True(t, s.LogRequestBody)
}

func TestWithLogRequestBody(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithLogRequestBody(true))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.True(t, s.LogRequestBody)
}

func TestLoadRequestIdModeDefault(t *testing.T) {
	t.Parallel()
	s, err := settings.Load()
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, "AcceptFromHeader", s.RequestIdMode.String())
}

func TestLoadRequestIdModeFromEnvironment(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithEnvGetter(mockEnvGetter("SIMBA_REQUEST_ID_MODE", "AcceptFromQuery")))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, "AcceptFromQuery", s.RequestIdMode.String())
}

func TestWithRequestIdMode(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithRequestIdMode(simbaModels.AlwaysGenerate))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, "AlwaysGenerate", s.RequestIdMode.String())
}

func TestLoadGenerateOpenAPIDocsDefault(t *testing.T) {
	t.Parallel()
	s, err := settings.Load()
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.True(t, s.GenerateOpenAPIDocs)
}

func TestLoadGenerateOpenAPIDocsFromEnvironment(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithEnvGetter(mockEnvGetter("SIMBA_DOCS_GENERATE_OPENAPI_DOCS", "true")))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.True(t, s.GenerateOpenAPIDocs)
}

func TestWithGenerateOpenAPIDocs(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithGenerateOpenAPIDocs(true))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.True(t, s.GenerateOpenAPIDocs)
}

func TestLoadMountDocsUIEndpointDefault(t *testing.T) {
	t.Parallel()
	s, err := settings.Load()
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.True(t, s.MountDocsUIEndpoint)
}

func TestLoadMountDocsUIEndpointFromEnvironment(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithEnvGetter(mockEnvGetter("SIMBA_DOCS_MOUNT_DOCS_UI_ENDPOINT", "false")))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.False(t, s.MountDocsUIEndpoint)
}

func TestWithMountDocsUIEndpoint(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithMountDocsUIEndpoint(false))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.False(t, s.MountDocsUIEndpoint)
}

func TestLoadOpenAPIFilePathDefault(t *testing.T) {
	t.Parallel()
	s, err := settings.Load()
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, "/openapi.json", s.OpenAPIFilePath)
}

func TestLoadOpenAPIFilePathFromEnvironment(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithEnvGetter(mockEnvGetter("SIMBA_DOCS_OPENAPI_FILE_PATH", "/api.json")))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, "/api.json", s.OpenAPIFilePath)
}

func TestWithOpenAPIFilePath(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithOpenAPIFilePath("/api.json"))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, "/api.json", s.OpenAPIFilePath)
}

func TestLoadDocsUIPathDefault(t *testing.T) {
	t.Parallel()
	s, err := settings.Load()
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, "/docs", s.DocsUIPath)
}

func TestLoadDocsUIPathFromEnvironment(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithEnvGetter(mockEnvGetter("SIMBA_DOCS_UI_PATH", "/api-docs")))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, "/api-docs", s.DocsUIPath)
}

func TestWithDocsUIPath(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithDocsUIPath("/api-docs"))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, "/api-docs", s.DocsUIPath)
}

func TestNilLogger(t *testing.T) {
	t.Parallel()
	s, err := settings.Load(settings.WithLogger(nil))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.NotNil(t, s.Logger)
}

func TestWithLogger(t *testing.T) {
	t.Parallel()
	customLogger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	s, err := settings.Load(settings.WithLogger(customLogger))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.NotNil(t, s.Logger)
}

func TestLoadWithOptions(t *testing.T) {
	t.Parallel()
	// Test that LoadWithOptions works the same as Load
	s, err := settings.LoadWithOptions(settings.WithServerPort(8080))
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, 8080, s.Port)
}

func mockEnvGetter(key, value string) func(key string) string {
	mockEnv := map[string]string{
		key: value,
	}

	getEnvFunc := func(key string) string {
		if val, ok := mockEnv[key]; ok {
			return val
		}
		return ""
	}

	return getEnvFunc
}
