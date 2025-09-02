package simba

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"
	"github.com/sillen102/simba/simbaErrors"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

var messageGenerators = map[string]func(validator.FieldError) string{
	// Comparisons
	"gte": func(e validator.FieldError) string {
		return fmt.Sprintf("%s must be greater than or equal to %s", strcase.ToLowerCamel(e.Field()), e.Param())
	},
	"lte": func(e validator.FieldError) string {
		return fmt.Sprintf("%s must be less than or equal to %s", strcase.ToLowerCamel(e.Field()), e.Param())
	},
	"gt": func(e validator.FieldError) string {
		return fmt.Sprintf("%s must be greater than %s", strcase.ToLowerCamel(e.Field()), e.Param())
	},
	"lt": func(e validator.FieldError) string {
		return fmt.Sprintf("%s must be less than %s", strcase.ToLowerCamel(e.Field()), e.Param())
	},

	// Strings
	"alpha": func(e validator.FieldError) string {
		return fmt.Sprintf("%s must contain only letters", strcase.ToLowerCamel(e.Field()))
	},
	"alphanum": func(e validator.FieldError) string {
		return fmt.Sprintf("%s must contain only letters and numbers", strcase.ToLowerCamel(e.Field()))
	},
	"alphanumunicode": func(e validator.FieldError) string {
		return fmt.Sprintf("%s must contain only letters and numbers that are part of unicode", strcase.ToLowerCamel(e.Field()))
	},
	"alphaunicode": func(e validator.FieldError) string {
		return fmt.Sprintf("%s must contain only letters (no numbers allowed) that are part of unicode", strcase.ToLowerCamel(e.Field()))
	},
	"number": func(e validator.FieldError) string {
		return fmt.Sprintf("%s must be a valid number", strcase.ToLowerCamel(e.Field()))
	},
	"numeric": func(e validator.FieldError) string {
		return fmt.Sprintf("%s must be a numeric value", strcase.ToLowerCamel(e.Field()))
	},

	// Format
	"base64": func(e validator.FieldError) string {
		return fmt.Sprintf("%s must be a valid base64 encoded string", strcase.ToLowerCamel(e.Field()))
	},
	"e164": func(e validator.FieldError) string {
		return fmt.Sprintf("'%s' must be a valid E.164 formatted phone number", getValueString(e))
	},
	"email": func(e validator.FieldError) string {
		return fmt.Sprintf("'%s' is not a valid email address", getValueString(e))
	},
	"jwt": func(e validator.FieldError) string {
		return fmt.Sprintf("%s must be a valid JWT token", strcase.ToLowerCamel(e.Field()))
	},
	"uuid": func(e validator.FieldError) string {
		return fmt.Sprintf("%s must be a valid UUID", strcase.ToLowerCamel(e.Field()))
	},

	// Other
	"len": func(e validator.FieldError) string {
		return fmt.Sprintf("%s must be exactly %s characters long", strcase.ToLowerCamel(e.Field()), e.Param())
	},
	"max": func(e validator.FieldError) string {
		if isNumeric(e.Kind()) {
			return fmt.Sprintf("%s must not exceed %s", strcase.ToLowerCamel(e.Field()), e.Param())
		} else if e.Kind() == reflect.Slice || e.Kind() == reflect.Array || e.Kind() == reflect.Map {
			return fmt.Sprintf("%s must not contain more than %s items", strcase.ToLowerCamel(e.Field()), e.Param())
		} else {
			return fmt.Sprintf("%s must not exceed %s characters", strcase.ToLowerCamel(e.Field()), e.Param())
		}
	},
	"min": func(e validator.FieldError) string {
		if isNumeric(e.Kind()) {
			return fmt.Sprintf("%s must be at least %s", strcase.ToLowerCamel(e.Field()), e.Param())
		} else if e.Kind() == reflect.Slice || e.Kind() == reflect.Array || e.Kind() == reflect.Map {
			return fmt.Sprintf("%s must contain at least %s items", strcase.ToLowerCamel(e.Field()), e.Param())
		} else {
			return fmt.Sprintf("%s must be at least %s characters long", strcase.ToLowerCamel(e.Field()), e.Param())
		}
	},
	"required": func(e validator.FieldError) string {
		return fmt.Sprintf("%s is required", strcase.ToLowerCamel(e.Field()))
	},
}

// Validator returns the validator instance for the application.
func Validator() *validator.Validate {
	return validate
}

// ValidateStruct is a helper function for validating requests using the validator
// package. If the request is nil, it will return nil. If the request is valid, it
// will return an empty slice of ValidationErrors. If the request is invalid, it
// will return a slice of ValidationErrors containing the validation errors for
// each field.
func ValidateStruct(request any) []string {
	if request == nil {
		return nil
	}

	err := validate.Struct(request)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return []string{"validation failed"}
	}

	if len(validationErrors) > 0 {
		validationErrorsData := make([]string, len(validationErrors))
		for i, e := range validationErrors {
			validationErrorsData[i] = mapValidationMessage(e)
		}
		return validationErrorsData
	}

	return nil
}

func mapValidationMessage(e validator.FieldError) string {
	// Look up the generator and call it, or use default
	if generator, exists := messageGenerators[e.Tag()]; exists {
		return generator(e)
	}
	return fmt.Sprintf("'%s' is invalid as input for %s", getValueString(e), strcase.ToLowerCamel(e.Field()))
}

func mapValidationErrors(validationErrors []string) *simbaErrors.SimbaError {
	var errorMessage string
	if len(validationErrors) == 1 {
		errorMessage = "request validation failed, 1 validation error"
	} else {
		errorMessage = fmt.Sprintf("request validation failed, %d validation errors", len(validationErrors))
	}

	return simbaErrors.NewSimbaError(
		http.StatusBadRequest,
		errorMessage,
		nil,
	).WithDetails(validationErrors)
}

func getValueString(e validator.FieldError) string {
	var valueStr string
	if str, ok := e.Value().(string); ok {
		valueStr = str
	} else {
		valueStr = fmt.Sprintf("%v", e.Value())
	}

	return valueStr
}

func isNumeric(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}
