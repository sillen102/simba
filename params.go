package simba

import (
	"encoding"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sillen102/simba/simbaErrors"
	"github.com/sillen102/simba/simbaModels"
)

// parseAndValidateParams creates a new instance of the parameter struct,
// populates it using the MapParams interface method, and validates it.
func parseAndValidateParams[Params any](r *http.Request) (Params, error) {
	var instance Params
	// If instance is NoParams or empty struct, return early
	if _, ok := any(instance).(simbaModels.NoParams); ok {
		return instance, nil
	}
	t := reflect.TypeOf(&instance).Elem()
	if t.NumField() == 0 {
		return instance, nil
	}
	v := reflect.ValueOf(&instance).Elem()

	validationErrors := make([]ValidationError, 0)

	// Extract parameters from struct tags and set values
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Handle embedded structs
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			embeddedInstance := fieldValue.Addr().Interface()
			if err := parseEmbeddedParams(r, embeddedInstance); err != nil {
				return instance, err
			}
			continue
		}

		if !fieldValue.CanSet() {
			continue
		}

		values := getParamValues(r, field)

		// If no values was provided, try to set default values
		if len(values) == 0 {
			if err := setDefaultValue(fieldValue, field); err != nil {
				// If the default values is not valid it's not a client error and should therefore return a 500
				return instance, simbaErrors.NewSimbaError(
					http.StatusInternalServerError,
					"invalid default values",
					err,
				).WithDetails(err.Error())
			}
			continue
		}

		if validationErr := setFieldValue(fieldValue, values, field); validationErr != nil {
			validationErrors = append(validationErrors, *validationErr)
		}
	}

	if len(validationErrors) == 0 {
		if valErrs := ValidateStruct(instance); len(valErrs) > 0 {
			validationErrors = append(validationErrors, valErrs...)
		}
	}

	if len(validationErrors) > 0 {
		return instance, simbaErrors.NewSimbaError(
			http.StatusBadRequest,
			"request validation failed",
			nil,
		).WithDetails(validationErrors)
	}

	return instance, nil
}

// parseEmbeddedParams processes embedded struct fields recursively
func parseEmbeddedParams(r *http.Request, embeddedInstance interface{}) error {
	t := reflect.TypeOf(embeddedInstance).Elem()
	v := reflect.ValueOf(embeddedInstance).Elem()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Handle nested embedded structs
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			if err := parseEmbeddedParams(r, fieldValue.Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		if !fieldValue.CanSet() {
			continue
		}

		values := getParamValues(r, field)

		// If no values were provided, try to set default values
		if len(values) == 0 {
			if err := setDefaultValue(fieldValue, field); err != nil {
				return simbaErrors.NewSimbaError(
					http.StatusInternalServerError,
					"invalid default values",
					err,
				).WithDetails(err.Error())
			}
			continue
		}

		if err := setFieldValue(fieldValue, values, field); err != nil {
			return err
		}
	}

	return nil
}

// getParamValues returns the parameter value based on the struct tag
func getParamValues(r *http.Request, field reflect.StructField) []string {
	switch {
	case field.Tag.Get("header") != "":
		return []string{r.Header.Get(field.Tag.Get("header"))}
	case field.Tag.Get("cookie") != "":
		cookie, err := r.Cookie(field.Tag.Get("cookie"))
		if err != nil {
			return nil
		}
		return []string{cookie.Value}
	case field.Tag.Get("path") != "":
		paramName := field.Tag.Get("path")
		return []string{r.PathValue(paramName)}
	case field.Tag.Get("query") != "":
		queryValues := r.URL.Query()[field.Tag.Get("query")]
		if len(queryValues) == 0 {
			return nil
		}
		// Split comma-separated values
		var result []string
		for _, value := range queryValues {
			result = append(result, strings.Split(value, ",")...)
		}
		return result
	}
	return nil
}

// getFieldName returns the parameter name from struct tags
func getFieldName(field reflect.StructField) string {
	if header := field.Tag.Get("header"); header != "" {
		return header
	} else if path := field.Tag.Get("path"); path != "" {
		return path
	} else if query := field.Tag.Get("query"); query != "" {
		return query
	} else if cookie := field.Tag.Get("cookie"); cookie != "" {
		return cookie
	} else if json := field.Tag.Get("json"); json != "" {
		return json
	}
	return field.Name
}

func setFieldValue(fieldValue reflect.Value, values []string, field reflect.StructField) *ValidationError {
	if len(values) == 0 {
		return nil
	}

	// Check if the field is a slice
	if fieldValue.Kind() == reflect.Slice {
		slice := reflect.MakeSlice(fieldValue.Type(), len(values), len(values))

		for i, value := range values {
			elem := slice.Index(i)
			if err := setSingleValue(elem, value, field); err != nil {
				return err
			}
		}

		fieldValue.Set(slice)
		return nil
	}

	// Handle single values
	if len(values) == 1 {
		return setSingleValue(fieldValue, values[0], field)
	}

	return &ValidationError{
		Field: getFieldName(field),
		Err:   fmt.Errorf("unsupported field type: %v", fieldValue.Kind()).Error(),
	}
}

// setSingleValue converts and sets a string value to the appropriate field type
func setSingleValue(fieldValue reflect.Value, value string, field reflect.StructField) *ValidationError {
	if value == "" {
		return nil
	}

	var err error
	switch fieldValue.Type().String() {
	case "time.Time":
		format := field.Tag.Get("format")
		if format == "" {
			format = time.RFC3339
		}
		var timeVal time.Time
		if timeVal, err = time.Parse(format, value); err != nil {
			return &ValidationError{
				Field: getFieldName(field),
				Err:   fmt.Errorf("invalid time parameter value: %s", value).Error(),
			}
		}
		fieldValue.Set(reflect.ValueOf(timeVal))
		return nil
	case "uuid.UUID":
		var uuidVal uuid.UUID
		if uuidVal, err = uuid.Parse(value); err != nil {
			return &ValidationError{
				Field: getFieldName(field),
				Err:   fmt.Errorf("invalid UUID parameter value: %s", value).Error(),
			}
		}
		fieldValue.Set(reflect.ValueOf(uuidVal))
		return nil
	}

	// Check if the type implements TextUnmarshaler (except time.Time and uuid.UUID which are handled separately)
	if fieldValue.CanAddr() {
		ptrVal := fieldValue.Addr()
		if unmarshaler, ok := ptrVal.Interface().(encoding.TextUnmarshaler); ok {
			if err = unmarshaler.UnmarshalText([]byte(value)); err != nil {
				fieldName := getFieldName(field)
				return &ValidationError{
					Field: fieldName,
					Err:   fmt.Errorf("invalid value %s for %s", value, fieldName).Error(),
				}
			}
			return nil
		}
	}

	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Int, reflect.Int64:
		var intVal int64
		if intVal, err = strconv.ParseInt(value, 10, 64); err != nil {
			return &ValidationError{
				Field: getFieldName(field),
				Err:   fmt.Errorf("invalid int parameter value: %s", value).Error(),
			}
		}
		fieldValue.SetInt(intVal)
		return nil
	case reflect.Bool:
		var boolVal bool
		if boolVal, err = strconv.ParseBool(value); err != nil {
			return &ValidationError{
				Field: getFieldName(field),
				Err:   fmt.Errorf("invalid bool parameter value: %s", value).Error(),
			}
		}
		fieldValue.SetBool(boolVal)
		return nil
	case reflect.Float64:
		var floatVal float64
		if floatVal, err = strconv.ParseFloat(value, 64); err != nil {
			return &ValidationError{
				Field: getFieldName(field),
				Err:   fmt.Errorf("invalid float parameter value: %s", value).Error(),
			}
		}
		fieldValue.SetFloat(floatVal)
		return nil
	default:
		return &ValidationError{
			Field: getFieldName(field),
			Err:   fmt.Errorf("unsupported field type: %v", fieldValue.Kind()).Error(),
		}
	}

	return nil
}

// setDefaultValue sets the default value from struct tag if available
func setDefaultValue(fieldValue reflect.Value, field reflect.StructField) *ValidationError {
	defaultVal := field.Tag.Get("default")
	if defaultVal == "" {
		return nil
	}
	// Split comma-separated values in case of slice
	defaultVals := strings.Split(defaultVal, ",")
	return setFieldValue(fieldValue, defaultVals, field)
}
