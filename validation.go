package simba

import (
	"errors"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

// ValidateStruct is a helper function for validating requests
func ValidateStruct(request any) ValidationErrors {
	err := validate.Struct(request)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		log.Printf("Validation failed with unexpected error: %v", err)
		return ValidationErrors{
			{
				Parameter: "unknown",
				Type:      "unknown",
				Message:   "An unknown validation error occurred.",
			},
		}
	}

	if len(validationErrors) > 0 {
		validationErrorsData := make(ValidationErrors, len(validationErrors))
		for i, e := range validationErrors {
			var valueStr string
			if str, ok := e.Value().(string); ok {
				valueStr = str
			} else {
				valueStr = fmt.Sprintf("%v", e.Value())
			}

			message := getValidationMessage(e, valueStr)
			validationErrorsData[i] = ValidationError{
				Parameter: strcase.ToLowerCamel(e.Field()),
				Type:      ParameterTypeBody,
				Message:   message,
			}
		}
		return validationErrorsData
	}

	return nil
}

// getValidationMessage returns appropriate error message based on the validation tag
func getValidationMessage(e validator.FieldError, value string) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", strcase.ToLowerCamel(e.Field()))
	case "email":
		return fmt.Sprintf("'%s' is not a valid email address", value)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", strcase.ToLowerCamel(e.Field()), e.Param())
	case "max":
		return fmt.Sprintf("%s must not exceed %s characters", strcase.ToLowerCamel(e.Field()), e.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", strcase.ToLowerCamel(e.Field()), e.Param())
	case "numeric":
		return fmt.Sprintf("'%s' must be a valid number", value)
	case "alphanum":
		return fmt.Sprintf("'%s' must contain only letters and numbers", value)
	default:
		return fmt.Sprintf("'%s' is invalid as input for %s", value, strcase.ToLowerCamel(e.Field()))
	}
}
