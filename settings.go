package simba

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/sillen102/simba/logging"
	"github.com/spf13/viper"
)

// Settings is a struct that holds the application Settings
type Settings struct {

	// Request Settings
	Request RequestSettings

	// Logging Settings
	Logging LoggingSettings
}

// RequestSettings holds the Settings for the Request processing
type RequestSettings struct {

	// UnknownFields will set the behavior for unknown fields in the Request body,
	// resulting in a 400 Bad Request response if a field is present that cannot be
	// mapped to the model struct.
	UnknownFields UnknownFields

	// LogRequestBody will determine if the Request body will be logged
	// If set to "disabled", the Request body will not be logged, which is also the default
	LogRequestBody zerolog.Level

	// AcceptRequestIdHeader will determine if the Request ID should be read from the
	// Request header. If not set, the Request ID will be generated.
	RequestId RequestId
}

// LoggingSettings holds the Settings for the logger
type LoggingSettings struct {

	// Level is the log Level for the logger that will be used
	Level zerolog.Level

	// Format is the log Format for the logger that will be used
	Format logging.LogFormat

	// Output is the Output for the logger that will be used
	Output io.Writer
}

type UnknownFields int

const (
	Allow UnknownFields = iota
	Disallow
)

func (u UnknownFields) String() string {
	switch u {
	case Allow:
		return "allow"
	case Disallow:
		return "disallow"
	default:
		return "disallow"
	}
}

func ParseUnknownFields(s string) UnknownFields {
	switch strings.ToLower(s) {
	case "disallow":
		return Disallow
	default:
		return Allow
	}
}

type RequestId int

const (
	AcceptFromHeader RequestId = iota
	Generate
)

func (r RequestId) String() string {
	switch r {
	case AcceptFromHeader:
		return "accept_from_header"
	case Generate:
		return "generate"
	default:
		return "accept_from_header"
	}
}

func ParseRequestId(s string) RequestId {
	switch strings.ToLower(s) {
	case "generate":
		return Generate
	default:
		return AcceptFromHeader
	}
}

func loadConfig(cfg ...Settings) (*Settings, error) {
	var config Settings
	if len(cfg) > 0 {
		config = cfg[0]
	}

	viper.AutomaticEnv()

	viper.SetDefault("REQUEST_UNKNOWN_FIELDS", Disallow.String())
	viper.SetDefault("REQUEST_ID", AcceptFromHeader.String())
	viper.SetDefault("LOG_REQUEST_BODY", "disabled")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORMAT", "text")
	viper.SetDefault("LOG_OUTPUT", "stdout")

	configFiles := []struct {
		name string
		typ  string
	}{
		{"config", "yml"},
		{"config", "yaml"},
		{"config", "json"},
		{".env", "env"},
	}

	configPaths := []string{
		".",        // Current directory
		"./config", // Config subdirectory
	}

	var err error
	for _, file := range configFiles {
		viper.SetConfigName(file.name)
		viper.SetConfigType(file.typ)

		// Clear any previously set paths
		viper.SetConfigPermissions(0)

		// Add all paths for this config type
		for _, path := range configPaths {
			viper.AddConfigPath(path)
		}

		if err = viper.ReadInConfig(); err == nil {
			break
		}
	}

	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, err
		}
	}

	// Only unmarshal into config if it's empty
	if config == (Settings{}) {
		if err = viper.Unmarshal(&config); err != nil {
			return nil, err
		}
	} else {
		// For non-empty config, only override unset fields
		if config.Request.UnknownFields == 0 {
			config.Request.UnknownFields = ParseUnknownFields(viper.GetString("REQUEST_UNKNOWN_FIELDS"))
		}
		if config.Request.RequestId == 0 {
			config.Request.RequestId = ParseRequestId(viper.GetString("REQUEST_ID"))
		}
		if config.Request.LogRequestBody == 0 {
			level, _ := zerolog.ParseLevel(viper.GetString("LOG_REQUEST_BODY"))
			config.Request.LogRequestBody = level
		}
		if config.Logging.Level == 0 {
			level, _ := zerolog.ParseLevel(viper.GetString("LOG_LEVEL"))
			config.Logging.Level = level
		}
		if config.Logging.Format == "" {
			config.Logging.Format = logging.LogFormat(viper.GetString("LOG_FORMAT"))
		}
	}

	// Handle log Output separately only if not already set
	if config.Logging.Output == nil {
		outputName := viper.GetString("LOG_OUTPUT")
		switch outputName {
		case "stdout":
			config.Logging.Output = os.Stdout
		case "stderr":
			config.Logging.Output = os.Stderr
		default:
			config.Logging.Output = os.Stdout
		}
	}

	return &config, nil
}
