package logging

import (
	"io"
	"os"
	"strings"

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
