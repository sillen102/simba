package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
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
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return fld.Name
		}
		return name
	})
	err := en_translations.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		panic("failed to register default translations for validator: " + err.Error())
	}
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
			validationErrorsData[i] = ValidationError{
				Field: e.Field(),
				Err:   e.Translate(trans),
			}
		}
		return validationErrorsData
	}

	return nil
}
