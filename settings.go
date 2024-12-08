package simba

import (
	"io"
	"os"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog"
	"github.com/sillen102/simba/middleware"
)

// TODO: Settings testing
// 	1. Configuration validation
//  2. Default settings behavior
//  3. Settings override scenarios
//  4. Invalid settings handling
//  5. Server settings

// Settings is a struct that holds the application Settings
type Settings struct {

	// Server settings
	Server ServerSettings

	// Request settings
	Request RequestSettings

	// Logging settings
	Logging LoggingConfig
}

// ServerSettings holds the Settings for the application server
type ServerSettings struct {

	// Host is the host the server will listen on
	Host string `yaml:"host" env:"HOST" env-default:"localhost"`

	// Addr is the address the server will listen on
	Port string `yaml:"port" env:"PORT" env-default:"8000"`
}

// RequestSettings holds the Settings for the Request processing
type RequestSettings struct {

	// AllowUnknownFields will set the behavior for unknown fields in the Request body,
	// resulting in a 400 Bad Request response if a field is present that cannot be
	// mapped to the model struct.
	AllowUnknownFields AllowUnknownFields `yaml:"request_allow_unknown_fields" env:"REQUEST_ALLOW_UNKNOWN_FIELDS" env-default:"disallow"`

	// LogRequestBody will determine if the Request body will be logged
	// If set to "disabled", the Request body will not be logged, which is also the default
	LogRequestBody zerolog.Level `yaml:"log_request_body" env:"LOG_REQUEST_BODY" env-default:"disabled"`

	// AcceptRequestIdHeader will determine if the Request ID should be read from the
	// Request header. If not set, the Request ID will be generated.
	RequestIdMode middleware.RequestIdMode `yaml:"request_id_mode" env:"REQUEST_ID_MODE" env-default:"accept_from_header"`
}

type AllowUnknownFields int

const (
	Allow AllowUnknownFields = iota
	Disallow
)

func (u AllowUnknownFields) String() string {
	switch u {
	case Allow:
		return "allow"
	case Disallow:
		return "disallow"
	default:
		return "disallow"
	}
}

func (u *AllowUnknownFields) SetValue(s string) error {
	*u = ParseAllowUnknownFields(s)
	return nil
}

func ParseAllowUnknownFields(s string) AllowUnknownFields {
	switch strings.ToLower(s) {
	case "disallow":
		return Disallow
	default:
		return Allow
	}
}

// LoggingConfig holds the settings for the logger
type LoggingConfig struct {

	// Level is the log Level for the logger that will be used
	Level zerolog.Level `yaml:"log_level" env:"LOG_LEVEL" env-default:"info"`

	// Format is the log Format for the logger that will be used
	Format LogFormat `yaml:"log_format" env:"LOG_FORMAT" env-default:"text"`

	// output is the name of the output for the logger that will be used
	output LogOutput `yaml:"log_output" env:"LOG_OUTPUT" env-default:"stdout"`

	// Output is the Output for the logger that will be used
	Output io.Writer
}

type LogFormat string

const (
	JsonFormat LogFormat = "json"
	TextFormat LogFormat = "text"
	TimeFormat string    = "2006-01-02T15:04:05.000000"
)

func (f LogFormat) String() string {
	switch f {
	case JsonFormat:
		return "json"
	default:
		return "text"
	}
}

func (f *LogFormat) SetValue(s string) error {
	*f = ParseLogFormat(s)
	return nil
}

func ParseLogFormat(s string) LogFormat {
	switch strings.ToLower(s) {
	case "json":
		return JsonFormat
	default:
		return TextFormat
	}
}

type LogOutput string

const (
	Stdout LogOutput = "stdout"
	Stderr LogOutput = "stderr"
)

func (o LogOutput) String() string {
	switch o {
	case Stdout:
		return "stdout"
	case Stderr:
		return "stderr"
	default:
		return "stdout"
	}
}

func ParseLogOutput(s string) LogOutput {
	switch strings.ToLower(s) {
	case "stdout":
		return Stdout
	case "stderr":
		return Stderr
	default:
		return Stdout
	}
}

func getOutput(o LogOutput) io.Writer {
	switch o {
	case Stdout:
		return os.Stdout
	case Stderr:
		return os.Stderr
	default:
		return os.Stdout
	}
}

func loadConfig(provided ...Settings) (*Settings, error) {
	var settings Settings
	err := cleanenv.ReadEnv(&settings)
	if err != nil {
		return nil, err
	}

	if len(provided) > 0 {
		st := provided[0]

		if st.Server.Host != "" {
			settings.Server.Host = st.Server.Host
		}

		if st.Server.Port != "" {
			settings.Server.Port = st.Server.Port
		}

		// Disallow is the default so if the user doesn't set any different, we don't override it
		if st.Request.AllowUnknownFields != Disallow {
			settings.Request.AllowUnknownFields = st.Request.AllowUnknownFields
		}

		// Disabled is the default so if the user doesn't set any different, we don't override it
		if st.Request.LogRequestBody != zerolog.Disabled {
			settings.Request.LogRequestBody = st.Request.LogRequestBody
		}

		// AcceptFromHeader is the default so if the user doesn't set any different, we don't override it
		if st.Request.RequestIdMode != middleware.AcceptFromHeader {
			settings.Request.RequestIdMode = st.Request.RequestIdMode
		}

		if st.Logging.Format != "" {
			settings.Logging.Format = st.Logging.Format
		}

		// Info is the default so if the user doesn't set any different, we don't override it
		if st.Logging.Level != zerolog.InfoLevel {
			settings.Logging.Level = st.Logging.Level
		}

		if st.Logging.Output != nil {
			settings.Logging.Output = st.Logging.Output
		}
	}

	return &settings, nil
}
