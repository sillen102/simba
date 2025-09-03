package simba

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/iancoleman/strcase"
)

type ValidationError struct {
	Field string `json:"field"`
	Err   string `json:"error"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("Validation failed on field '%s': %s", e.Field, e.Err)
}

var (
	uni      *ut.UniversalTranslator
	trans    ut.Translator
	validate *validator.Validate
)

func init() {
	enLocale := en.New()
	uni = ut.New(enLocale, enLocale)
	trans, _ = uni.GetTranslator("en")

	validate = validator.New(validator.WithRequiredStructEnabled())
	en_translations.RegisterDefaultTranslations(validate, trans)

	// Register custom translations for each tag

	// Comparisons
	_ = validate.RegisterTranslation("gte", trans,
		func(ut ut.Translator) error {
			return ut.Add("gte", "{0} must be greater than or equal to {1}", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			field := strcase.ToLowerCamel(fe.Field())
			return fmt.Sprintf("%s must be greater than or equal to %s", field, fe.Param())
		},
	)

	_ = validate.RegisterTranslation("lte", trans,
		func(ut ut.Translator) error {
			return ut.Add("lte", "{0} must be less than or equal to {1}", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			field := strcase.ToLowerCamel(fe.Field())
			return fmt.Sprintf("%s must be less than or equal to %s", field, fe.Param())
		},
	)

	_ = validate.RegisterTranslation("gt", trans,
		func(ut ut.Translator) error {
			return ut.Add("gt", "{0} must be greater than {1}", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			field := strcase.ToLowerCamel(fe.Field())
			return fmt.Sprintf("%s must be greater than %s", field, fe.Param())
		},
	)

	_ = validate.RegisterTranslation("lt", trans,
		func(ut ut.Translator) error {
			return ut.Add("lt", "{0} must be less than {1}", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			field := strcase.ToLowerCamel(fe.Field())
			return fmt.Sprintf("%s must be less than %s", field, fe.Param())
		},
	)

	// Strings
	_ = validate.RegisterTranslation("alpha", trans,
		func(ut ut.Translator) error {
			return ut.Add("alpha", "{0} must contain only letters", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			field := strcase.ToLowerCamel(fe.Field())
			return fmt.Sprintf("%s must contain only letters", field)
		},
	)

	_ = validate.RegisterTranslation("alphanum", trans,
		func(ut ut.Translator) error {
			return ut.Add("alphanum", "{0} must contain only letters and numbers", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			field := strcase.ToLowerCamel(fe.Field())
			return fmt.Sprintf("%s must contain only letters and numbers", field)
		},
	)

	_ = validate.RegisterTranslation("alphanumunicode", trans,
		func(ut ut.Translator) error {
			return ut.Add("alphanumunicode", "{0} must contain only letters and numbers that are part of unicode", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			field := strcase.ToLowerCamel(fe.Field())
			return fmt.Sprintf("%s must contain only letters and numbers that are part of unicode", field)
		},
	)

	_ = validate.RegisterTranslation("alphaunicode", trans,
		func(ut ut.Translator) error {
			return ut.Add("alphaunicode", "{0} must contain only letters (no numbers allowed) that are part of unicode", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			field := strcase.ToLowerCamel(fe.Field())
			return fmt.Sprintf("%s must contain only letters (no numbers allowed) that are part of unicode", field)
		},
	)

	_ = validate.RegisterTranslation("number", trans,
		func(ut ut.Translator) error {
			return ut.Add("number", "{0} must be a valid number", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			field := strcase.ToLowerCamel(fe.Field())
			return fmt.Sprintf("%s must be a valid number", field)
		},
	)

	_ = validate.RegisterTranslation("numeric", trans,
		func(ut ut.Translator) error {
			return ut.Add("numeric", "{0} must be a numeric value", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			field := strcase.ToLowerCamel(fe.Field())
			return fmt.Sprintf("%s must be a numeric value", field)
		},
	)

	// Format
	_ = validate.RegisterTranslation("base64", trans,
		func(ut ut.Translator) error {
			return ut.Add("base64", "{0} must be a valid base64 encoded string", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			field := strcase.ToLowerCamel(fe.Field())
			return fmt.Sprintf("%s must be a valid base64 encoded string", field)
		},
	)

	_ = validate.RegisterTranslation("e164", trans,
		func(ut ut.Translator) error {
			return ut.Add("e164", "'{0}' must be a valid E.164 formatted phone number", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			return fmt.Sprintf("'%s' must be a valid E.164 formatted phone number", getValueString(fe))
		},
	)

	_ = validate.RegisterTranslation("email", trans,
		func(ut ut.Translator) error {
			return ut.Add("email", "'{0}' is not a valid email address", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			return fmt.Sprintf("'%s' is not a valid email address", getValueString(fe))
		},
	)

	_ = validate.RegisterTranslation("jwt", trans,
		func(ut ut.Translator) error {
			return ut.Add("jwt", "{0} must be a valid JWT token", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			field := strcase.ToLowerCamel(fe.Field())
			return fmt.Sprintf("%s must be a valid JWT token", field)
		},
	)

	_ = validate.RegisterTranslation("uuid", trans,
		func(ut ut.Translator) error {
			return ut.Add("uuid", "{0} must be a valid UUID", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			field := strcase.ToLowerCamel(fe.Field())
			return fmt.Sprintf("%s must be a valid UUID", field)
		},
	)

	// Other
	_ = validate.RegisterTranslation("len", trans,
		func(ut ut.Translator) error {
			return ut.Add("len", "{0} must be exactly {1} characters long", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			field := strcase.ToLowerCamel(fe.Field())
			return fmt.Sprintf("%s must be exactly %s characters long", field, fe.Param())
		},
	)

	_ = validate.RegisterTranslation("max", trans,
		func(ut ut.Translator) error {
			return ut.Add("max", "{0} must not exceed {1}", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			field := strcase.ToLowerCamel(fe.Field())
			if isNumeric(fe.Kind()) {
				return fmt.Sprintf("%s must not exceed %s", field, fe.Param())
			} else if fe.Kind() == reflect.Slice || fe.Kind() == reflect.Array || fe.Kind() == reflect.Map {
				return fmt.Sprintf("%s must not contain more than %s items", field, fe.Param())
			} else {
				return fmt.Sprintf("%s must not exceed %s characters", field, fe.Param())
			}
		},
	)

	_ = validate.RegisterTranslation("min", trans,
		func(ut ut.Translator) error {
			return ut.Add("min", "{0} must be at least {1}", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			field := strcase.ToLowerCamel(fe.Field())
			if isNumeric(fe.Kind()) {
				return fmt.Sprintf("%s must be at least %s", field, fe.Param())
			} else if fe.Kind() == reflect.Slice || fe.Kind() == reflect.Array || fe.Kind() == reflect.Map {
				return fmt.Sprintf("%s must contain at least %s items", field, fe.Param())
			} else {
				return fmt.Sprintf("%s must be at least %s characters long", field, fe.Param())
			}
		},
	)

	_ = validate.RegisterTranslation("required", trans,
		func(ut ut.Translator) error {
			return ut.Add("required", "{0} is required", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			field := strcase.ToLowerCamel(fe.Field())
			return fmt.Sprintf("%s is required", field)
		},
	)
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
func ValidateStruct(request any) []ValidationError {
	if request == nil {
		return nil
	}

	err := validate.Struct(request)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return []ValidationError{{Field: "unknown", Err: "validation failed"}}
	}

	if len(validationErrors) > 0 {
		validationErrorsData := make([]ValidationError, len(validationErrors))
		for i, e := range validationErrors {
			validationErrorsData[i] = MapValidationError(e, request)
		}
		return validationErrorsData
	}

	return nil
}

func MapValidationError(err validator.FieldError, request any) ValidationError {
	fieldName := err.StructField()
	typ := reflect.TypeOf(request)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	field, ok := typ.FieldByName(fieldName)
	fieldNameOut := ""
	if ok {
		fieldNameOut = getFieldName(field)
	} else {
		fieldNameOut = err.Field()
	}

	return ValidationError{
		Field: fieldNameOut,
		Err:   err.Translate(trans),
	}
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
