package settings

import (
	"errors"
	"log/slog"
	"reflect"
	"strconv"

	"github.com/sillen102/simba/enums"
)

// Config is a struct that holds the application settings
type Config struct {

	// Server settings
	Server

	// Request settings
	Request

	// Docs settings
	Docs

	// Logger settings
	Logger *slog.Logger
}

// Server holds the Config for the application server
type Server struct {

	// Host is the host the server will listen on
	Host string `default:"0.0.0.0"`

	// Addr is the address the server will listen on
	Port int `default:"9999"`
}

// Request holds the Config for the Request processing
type Request struct {

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

type Docs struct {

	// GenerateOpenAPIDocs will determine if the API documentation will be generated
	GenerateOpenAPIDocs bool `default:"true"`

	// MountDocsEndpoint will determine if the API documentation endpoint will be mounted
	MountDocsEndpoint bool `default:"true"`

	// DocsPath is the path to the API documentation
	DocsPath string `default:"/docs"`

	// ServiceName is the name of the service
	ServiceName string `default:"Simba Application"`
}

func Load(st ...Config) (*Config, error) {
	var settings Config
	if err := setDefaults(&settings); err != nil {
		return nil, err
	}

	if len(st) > 0 {
		provided := st[0]
		providedVal := reflect.ValueOf(provided)
		settingsVal := reflect.ValueOf(&settings).Elem()
		defaultVal := reflect.ValueOf(Config{})

		for i := 0; i < providedVal.NumField(); i++ {
			providedField := providedVal.Field(i)
			settingsField := settingsVal.Field(i)
			defaultField := defaultVal.Field(i)

			if !providedField.IsZero() && !reflect.DeepEqual(providedField.Interface(), defaultField.Interface()) {
				settingsField.Set(providedField)
			}
		}

		if provided.Logger != nil {
			settings.Logger = provided.Logger
		}
	}

	if settings.Logger == nil {
		settings.Logger = slog.Default()
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
			case reflect.Bool:
				boolValue, err := strconv.ParseBool(defaultTag)
				if err != nil {
					return err
				}
				field.SetBool(boolValue)
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
