package simba

import (
	"errors"
	"log/slog"
	"os"
	"reflect"
	"strconv"

	"github.com/sillen102/simba/enums"
	"github.com/sillen102/simba/logging"
)

// Settings is a struct that holds the application Settings
type Settings struct {

	// Server settings
	Server ServerSettings

	// Request settings
	Request RequestSettings

	// Logging settings
	Logging logging.Config
}

// ServerSettings holds the Settings for the application server
type ServerSettings struct {

	// Host is the host the server will listen on
	Host string `default:"0.0.0.0"`

	// Addr is the address the server will listen on
	Port int `default:"9999"`
}

// RequestSettings holds the Settings for the Request processing
type RequestSettings struct {

	// AllowUnknownFields will set the behavior for unknown fields in the Request body,
	// resulting in a 400 Bad Request response if a field is present that cannot be
	// mapped to the model struct.
	AllowUnknownFields enums.AllowOrNot `default:"Disallow"`

	// LogRequestBody will determine if the Request body will be logged
	// If set to "disabled", the Request body will not be logged, which is also the default
	LogRequestBody enums.EnableDisable `default:"Disabled"`

	// RequestIdMode determines how the Request ID will be handled
	RequestIdMode enums.RequestIdMode `default:"AcceptFromHeader"`
}

func loadConfig(st ...Settings) (*Settings, error) {
	var settings Settings
	if err := setDefaults(&settings); err != nil {
		return nil, err
	}
	settings.Logging.Output = os.Stdout

	if len(st) > 0 {
		provided := st[0]

		if provided.Server.Host != "" {
			settings.Server.Host = provided.Server.Host
		}

		if provided.Server.Port != 0 {
			settings.Server.Port = provided.Server.Port
		}

		// Disallow is the default so if the user doesn't set any different, we don't override it
		if provided.Request.AllowUnknownFields != enums.Disallow {
			settings.Request.AllowUnknownFields = provided.Request.AllowUnknownFields
		}

		// Disabled is the default so if the user doesn't set any different, we don't override it
		if provided.Request.LogRequestBody != enums.Disabled {
			settings.Request.LogRequestBody = provided.Request.LogRequestBody
		}

		// AcceptFromHeader is the default so if the user doesn't set any different, we don't override it
		if provided.Request.RequestIdMode != enums.AcceptFromHeader {
			settings.Request.RequestIdMode = provided.Request.RequestIdMode
		}

		if provided.Logging.Format != "" {
			settings.Logging.Format = provided.Logging.Format
		}

		// Info is the default so if the user doesn't set any different, we don't override it
		if provided.Logging.Level != slog.LevelInfo {
			settings.Logging.Level = provided.Logging.Level
		}

		if provided.Logging.Output != nil {
			settings.Logging.Output = provided.Logging.Output
		}
	}

	return &settings, nil
}

func setDefaults(ptr interface{}) error {
	val := reflect.ValueOf(ptr)
	if val.Kind() != reflect.Pointer || val.Elem().Kind() != reflect.Struct {
		return errors.New("provided argument must be a pointer to a struct")
	}

	val = val.Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if field.Kind() == reflect.Struct {
			if err := setDefaults(field.Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		defaultTag := fieldType.Tag.Get("default")
		if defaultTag == "" {
			continue // Skip if no default tag is present
		}

		// If the field can be set, update it
		if field.CanSet() {
			switch field.Kind() {
			case reflect.String:
				field.SetString(defaultTag)
			case reflect.Int, reflect.Int64:
				switch field.Type() {
				case reflect.TypeOf(enums.AllowOrNot(0)):
					// Handle AllowOrNot enum
					if defaultTag == enums.Allow.String() {
						field.SetInt(int64(enums.Allow))
					} else {
						field.SetInt(int64(enums.Disallow))
					}
				case reflect.TypeOf(enums.RequestIdMode(0)):
					// Handle RequestIdMode enum
					if defaultTag == enums.AcceptFromHeader.String() {
						field.SetInt(int64(enums.AcceptFromHeader))
					} else {
						field.SetInt(int64(enums.AlwaysGenerate))
					}
				case reflect.TypeOf(enums.EnableDisable(0)):
					// Handle EnableDisable enum
					if defaultTag == enums.Enabled.String() {
						field.SetInt(int64(enums.Enabled))
					} else {
						field.SetInt(int64(enums.Disabled))
					}
				case reflect.TypeOf(slog.Level(0)):
					// Handle slog.Level enum
					level, err := logging.ParseLogLevel(defaultTag)
					if err != nil {
						return err
					}
					field.SetInt(int64(level))
				default:
					intValue, err := strconv.Atoi(defaultTag)
					if err != nil {
						return err
					}
					field.SetInt(int64(intValue))
				}
			default:
				continue
			}
		}
	}

	return nil
}
