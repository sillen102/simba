package simba

import (
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog"
	"github.com/sillen102/simba/logging"
	"github.com/sillen102/simba/middleware"
)

// Settings is a struct that holds the application Settings
type Settings struct {

	// Request settings
	Request RequestSettings

	// Logging settings
	Logging logging.Config
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

func (f *AllowUnknownFields) SetValue(s string) error {
	*f = ParseAllowUnknownFields(s)
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

func loadConfig(provided ...Settings) (*Settings, error) {
	var settings Settings
	err := cleanenv.ReadEnv(&settings)
	if err != nil {
		return nil, err
	}

	if len(provided) > 0 {
		st := provided[0]

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
