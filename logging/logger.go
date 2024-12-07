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

var (
	defaultLogger zerolog.Logger
	once          sync.Once
)

// Init initializes the logger for the application with the provided config
func Init(config Config) {
	if config.Output == nil {
		config.Output = getOutput(config.output)
	}

	var logger zerolog.Logger
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
		zerolog.TimestampFunc = func() time.Time {
			return time.Now().UTC()
		}
		zerolog.TimeFieldFormat = TimeFormat
		defaultLogger = logger
		zerolog.DefaultContextLogger = &defaultLogger
	})
}

// With returns a new context with the provided logger or with the default logger if no logger is provided
func With(ctx context.Context, logger ...zerolog.Logger) context.Context {
	if len(logger) > 0 {
		return logger[0].WithContext(ctx)
	}
	return defaultLogger.WithContext(ctx)
}

// Get returns the logger from the context
// Returns the default logger if no logger is found in the context or no context is provided
func Get(ctx ...context.Context) *zerolog.Logger {
	if len(ctx) == 0 {
		return &defaultLogger
	}

	logger := zerolog.Ctx(ctx[0])
	if logger == nil || logger.GetLevel() == zerolog.Disabled {
		return &defaultLogger
	}

	return logger
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
