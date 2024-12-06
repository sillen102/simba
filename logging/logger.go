package logging

import (
	"context"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Config holds the settings for the logger
type Config struct {

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

var (
	defaultLogger zerolog.Logger
	once          sync.Once
)

func init() {
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}
	zerolog.TimeFieldFormat = TimeFormat
}

// New creates a new logger with the provided configuration
func New(config Config) zerolog.Logger {
	var logger zerolog.Logger

	if config.Output == nil {
		config.Output = getOutput(config.output)
	}

	if config.Format == JsonFormat {
		logger = zerolog.New(config.Output).Level(config.Level).With().Timestamp().Logger()
	} else {
		logger = zerolog.New(config.Output).Level(config.Level).Output(zerolog.ConsoleWriter{
			Out:          config.Output,
			TimeLocation: time.UTC,
			TimeFormat:   TimeFormat,
		}).With().Timestamp().Logger()
	}
	once.Do(func() {
		defaultLogger = logger
		zerolog.DefaultContextLogger = &logger
	})
	return logger
}

// WithLogger returns a new context with the provided logger
func WithLogger(ctx context.Context, logger zerolog.Logger) context.Context {
	return logger.WithContext(ctx)
}

// FromCtx returns the logger from the context
func FromCtx(ctx context.Context) *zerolog.Logger {
	logger := zerolog.Ctx(ctx)
	if logger == nil || logger.GetLevel() == zerolog.Disabled {
		return &defaultLogger
	}
	return logger
}

// GetDefault returns the default logger
func GetDefault() *zerolog.Logger {
	return &defaultLogger
}
