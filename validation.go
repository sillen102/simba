package simba

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"
	"github.com/sillen102/simba/simbaErrors"
)

// TODO: Validation testing
// 	1. Custom validation messages generation
// 	2. Edge cases with various data types
// 	3. Error handling for invalid structs

var validate = validator.New(validator.WithRequiredStructEnabled())

// ValidateStruct is a helper function for validating requests using the validator
// package. If the request is nil, it will return nil. If the request is valid, it
// will return an empty slice of ValidationErrors. If the request is invalid, it
// will return a slice of ValidationErrors containing the validation errors for
// each field.
func ValidateStruct(request any, paramType simbaErrors.ParameterType) simbaErrors.ValidationErrors {
	if request == nil {
		return nil
	}

	err := validate.Struct(request)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return simbaErrors.ValidationErrors{
			{
				Parameter: "unknown",
				Type:      paramType,
				Message:   "validation failed",
			},
		}
	}

	if len(validationErrors) > 0 {
		validationErrorsData := make(simbaErrors.ValidationErrors, len(validationErrors))
		for i, e := range validationErrors {
			var valueStr string
			if str, ok := e.Value().(string); ok {
				valueStr = str
			} else {
				valueStr = fmt.Sprintf("%v", e.Value())
			}

			message := MapValidationMessage(e, valueStr)
			validationErrorsData[i] = simbaErrors.ValidationError{
				Parameter: mapFieldName(e),
				Type:      paramType,
				Message:   message,
			}
		}
		return validationErrorsData
	}

	return nil
}

// MapValidationMessage returns appropriate error message based on the validation tag
func MapValidationMessage(e validator.FieldError, value string) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", strcase.ToLowerCamel(e.Field()))
	case "email":
		return fmt.Sprintf("'%s' is not a valid email address", value)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", strcase.ToLowerCamel(e.Field()), e.Param())
	case "max":
		return fmt.Sprintf("%s must not exceed %s characters", strcase.ToLowerCamel(e.Field()), e.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", strcase.ToLowerCamel(e.Field()), e.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", strcase.ToLowerCamel(e.Field()), e.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", strcase.ToLowerCamel(e.Field()), e.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", strcase.ToLowerCamel(e.Field()), e.Param())
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

func mapFieldName(e validator.FieldError) string {
	// Get the full namespace (e.g., "CreateUserRequest.Address.PostalCode")
	ns := e.Namespace()

	// Find the first dot position to skip the root struct name
	firstDotPos := -1
	for i, char := range ns {
		if char == '.' {
			firstDotPos = i
			break
		}
	}

	// If there's no dot or it's the last character, return the field name directly
	if firstDotPos == -1 || firstDotPos >= len(ns)-1 {
		return strcase.ToLowerCamel(e.Field())
	}

	// Extract everything after the first dot
	fieldPath := ns[firstDotPos+1:]

	// Split the field path by dots
	parts := strings.Split(fieldPath, ".")

	// Convert each part to lowerCamelCase
	for i, part := range parts {
		parts[i] = strcase.ToLowerCamel(part)
	}

	// Join the parts back with dots
	return strings.Join(parts, ".")
}
