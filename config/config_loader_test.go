package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/sillen102/simba/config"
)

type TestConfig struct {
	TestString string  `env:"TEST_STRING" default:"test"`
	TestInt    int     `env:"TEST_INT" default:"123"`
	TestInt64  int64   `env:"TEST_INT64" default:"123456"`
	TestBool   bool    `env:"TEST_BOOL" default:"true"`
	TestFloat  float64 `env:"TEST_FLOAT" default:"1.23"`
	Nested     NestedConfig
}

type NestedConfig struct {
	TestString string `env:"NESTED_TEST_STRING" default:"test"`
}

func TestLoadYamlFile(t *testing.T) {
	yamlContent := `
test:
  string: test2
  int: 456
  int64: 456789
  bool: false
  float: 1.45
nested:
  test:
    string: nested_test
`

	yamlFilePath, cleanup := createTempFile(t, yamlContent)
	defer cleanup()

	cfg := TestConfig{}
	loader := config.NewConfigLoader(&config.ConfigLoaderOpts{
		FilePath: yamlFilePath,
	})

	err := loader.Load(&cfg)
	if err != nil {
		t.Fatalf("Failed to load YAML file: %v", err)
	}

	expected := TestConfig{
		TestString: "test2",
		TestInt:    456,
		TestInt64:  456789,
		TestBool:   false,
		TestFloat:  1.45,
		Nested: NestedConfig{
			TestString: "nested_test",
		},
	}

	if !reflect.DeepEqual(cfg, expected) {
		t.Errorf("Expected %+v, got %+v", expected, cfg)
	}
}

func TestYamlFileWithEnvironmentOverride(t *testing.T) {
	yamlContent := `
test:
  string: from_yaml
  int: 123
  float: 1.23
nested:
  test:
    string: nested_from_yaml
`

	yamlFilePath, cleanup := createTempFile(t, yamlContent)
	defer cleanup()

	// Set environment variable to override YAML value
	_ = os.Setenv("TEST_STRING", "from_env")
	defer func() {
		_ = os.Unsetenv("TEST_STRING")
	}()

	cfg := TestConfig{}
	loader := config.NewConfigLoader(&config.ConfigLoaderOpts{
		FilePath: yamlFilePath,
	})

	err := loader.Load(&cfg)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Environment variable should override YAML value
	if cfg.TestString != "from_env" {
		t.Errorf("Expected TestString to be 'from_env', got '%s'", cfg.TestString)
	}

	// YAML value should be used for non-overridden field
	if cfg.Nested.TestString != "nested_from_yaml" {
		t.Errorf("Expected Nested.TestString to be 'nested_from_yaml', got '%s'", cfg.Nested.TestString)
	}
}

func TestFileExtensionPriority(t *testing.T) {
	// Test with both YAML and ENV content in different files
	yamlContent := `
test:
  string: from_yaml`
	envContent := `TEST_STRING=from_env`

	// Create both files
	yamlFilePath, yamlCleanup := createTempFile(t, yamlContent)
	defer yamlCleanup()

	envFilePath, envCleanup := createTempFile(t, envContent)
	defer envCleanup()

	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "yaml file",
			filePath: yamlFilePath,
			expected: "from_yaml",
		},
		{
			name:     "env file",
			filePath: envFilePath,
			expected: "from_env",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := TestConfig{}
			loader := config.NewConfigLoader(&config.ConfigLoaderOpts{
				FilePath: tc.filePath,
			})

			err := loader.Load(&cfg)
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			if cfg.TestString != tc.expected {
				t.Errorf("Expected TestString to be '%s', got '%s'", tc.expected, cfg.TestString)
			}
		})
	}
}

func TestInvalidYamlFallbackToEnv(t *testing.T) {
	// Create a file with valid ENV format but invalid YAML
	content := `
TEST_STRING=from_env
TEST_INT=456
# This is a comment that would break YAML parsing
`
	filePath, cleanup := createTempFile(t, content)
	defer cleanup()

	cfg := TestConfig{}
	loader := config.NewConfigLoader(&config.ConfigLoaderOpts{
		FilePath: filePath,
	})

	err := loader.Load(&cfg)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.TestString != "from_env" {
		t.Errorf("Expected TestString to be 'from_env', got '%s'", cfg.TestString)
	}

	if cfg.TestInt != 456 {
		t.Errorf("Expected TestInt to be 456, got %d", cfg.TestInt)
	}
}

func TestLoadEnvVars(t *testing.T) {

	cfg := TestConfig{}
	loader := config.NewConfigLoader(nil)

	tests := []struct {
		name       string
		setEnvVars func()
		expected   TestConfig
	}{
		{
			name: "load default values",
			expected: TestConfig{
				TestString: "test",
				TestInt:    123,
				TestInt64:  123456,
				TestBool:   true,
				TestFloat:  1.23,
				Nested: NestedConfig{
					TestString: "test",
				},
			},
		},
		{
			name: "load from env vars",
			setEnvVars: func() {
				_ = os.Setenv("TEST_STRING", "test2")
				_ = os.Setenv("TEST_INT", "456")
				_ = os.Setenv("TEST_INT64", "456789")
				_ = os.Setenv("TEST_BOOL", "false")
				_ = os.Setenv("TEST_FLOAT", "1.45")
				_ = os.Setenv("NESTED_TEST_STRING", "nested_test")
			},
			expected: TestConfig{
				TestString: "test2",
				TestInt:    456,
				TestInt64:  456789,
				TestBool:   false,
				TestFloat:  1.45,
				Nested: NestedConfig{
					TestString: "nested_test",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setEnvVars != nil {
				tc.setEnvVars()
			}

			_ = loader.Load(&cfg)
			if !reflect.DeepEqual(cfg, tc.expected) {
				t.Errorf("Expected %+v, got %+v", tc.expected, cfg)
			}
		})
	}
}

func TestLoadEnvFromDockerFile(t *testing.T) {

	type testStruct struct {
		TestString string `env:"FILE_TEST_STRING" default:"test"`
	}

	dockerSecretContent := "fileTest123"
	dockerSecretPath, cleanup := createTempFile(t, dockerSecretContent)
	defer cleanup()

	_ = os.Setenv("FILE_TEST_STRING_FILE", dockerSecretPath)

	cfg := testStruct{}
	loader := config.NewConfigLoader(nil)
	_ = loader.Load(&cfg)

	if cfg.TestString != dockerSecretContent {
		t.Errorf("Expected %s, got %s", dockerSecretContent, cfg.TestString)
	}
}

func TestLoadEnvFile(t *testing.T) {

	envContent := `
TEST_STRING=test2
TEST_INT=456
TEST_INT64=456789
TEST_BOOL=false
TEST_FLOAT=1.45
NESTED_TEST_STRING=nested_test
`

	envFilePath, cleanup := createTempFile(t, envContent)
	defer cleanup()

	cfg := TestConfig{}
	loader := config.NewConfigLoader(&config.ConfigLoaderOpts{
		FilePath: envFilePath,
	})

	tests := []struct {
		name     string
		expected TestConfig
	}{
		{
			name: "load from env file",
			expected: TestConfig{
				TestString: "test2",
				TestInt:    456,
				TestInt64:  456789,
				TestBool:   false,
				TestFloat:  1.45,
				Nested: NestedConfig{
					TestString: "nested_test",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_ = loader.Load(&cfg)
			if !reflect.DeepEqual(cfg, tc.expected) {
				t.Errorf("Expected %+v, got %+v", tc.expected, cfg)
			}
		})
	}
}

func createTempFile(t *testing.T, content string) (string, func()) {
	t.Helper()

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Determine file extension based on content
	var filename string
	if strings.Contains(content, ":") && !strings.Contains(content, "=") {
		filename = "config.yaml"
	} else {
		filename = ".env"
	}

	// Create file in the temp directory
	filePath := filepath.Join(tmpDir, filename)
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("Failed to write file: %v", err)
	}

	// Return the path and a cleanup function
	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	return filePath, cleanup
}
