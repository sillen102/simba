package simba

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// parseAndValidateParams creates a new instance of the parameter struct,
// populates it using the MapParams interface method, and validates it.
func parseAndValidateParams[Params any](r *http.Request) (Params, error) {
	var instance Params
	// If instance is NoParams or empty struct, return early
	if _, ok := any(instance).(NoParams); ok {
		return instance, nil
	}
	t := reflect.TypeOf(&instance).Elem()
	if t.NumField() == 0 {
		return instance, nil
	}
	v := reflect.ValueOf(&instance).Elem()

	// Extract parameters from struct tags and set values
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		// Get value from Request
		value := getParamValue(r, field)

		// If no value was provided, try to set default value
		if value == "" {
			if err := setDefaultValue(fieldValue, field); err != nil {
				return instance, NewHttpError(http.StatusBadRequest, "invalid default value", err,
					ValidationError{Parameter: field.Name, Type: getParamType(field), Message: "invalid default value"})
			}
			continue
		}

		// Set field value
		if err := setFieldValue(fieldValue, value); err != nil {
			return instance, NewHttpError(http.StatusBadRequest, err.Error(), err,
				ValidationError{Parameter: field.Name, Type: getParamType(field), Message: "invalid parameter value: " + value})
		}
	}

	// Validate required fields
	if validationErrors := validateStruct(instance, getParamType(t.Field(0))); len(validationErrors) > 0 {
		return instance, NewHttpError(http.StatusBadRequest, "request validation failed", nil, validationErrors...)
	}

	return instance, nil
}

// getParamValue returns the parameter value based on the struct tag
func getParamValue(r *http.Request, field reflect.StructField) string {
	switch {
	case field.Tag.Get("header") != "":
		return r.Header.Get(field.Tag.Get("header"))
	case field.Tag.Get("path") != "":
		if params := httprouter.ParamsFromContext(r.Context()); params != nil {
			return params.ByName(field.Tag.Get("path"))
		}
	case field.Tag.Get("query") != "":
		return r.URL.Query().Get(field.Tag.Get("query"))
	}
	return ""
}

// getParamType returns the parameter type based on the struct tag
func getParamType(field reflect.StructField) ParameterType {
	switch {
	case field.Tag.Get("header") != "":
		return ParameterTypeHeader
	case field.Tag.Get("path") != "":
		return ParameterTypePath
	case field.Tag.Get("query") != "":
		return ParameterTypeQuery
	default:
		return ParameterTypeBody
	}
}

// setFieldValue converts and sets a string value to the appropriate field type
func setFieldValue(fieldValue reflect.Value, value string) error {
	if value == "" {
		return nil
	}

	var err error
	switch fieldValue.Type().String() {
	case "time.Time":
		var timeVal time.Time
		if timeVal, err = time.Parse(time.RFC3339, value); err != nil {
			return fmt.Errorf("invalid time parameter value: %s", value)
		}
		fieldValue.Set(reflect.ValueOf(timeVal))
		return nil
	case "uuid.UUID":
		var uuidVal uuid.UUID
		if uuidVal, err = uuid.Parse(value); err != nil {
			return fmt.Errorf("invalid UUID parameter value: %s", value)
		}
		fieldValue.Set(reflect.ValueOf(uuidVal))
		return nil
	}

	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Int, reflect.Int64:
		var intVal int64
		if intVal, err = strconv.ParseInt(value, 10, 64); err != nil {
			return fmt.Errorf("invalid int parameter value: %s", value)
		}
		fieldValue.SetInt(intVal)
		return nil
	case reflect.Bool:
		var boolVal bool
		if boolVal, err = strconv.ParseBool(value); err != nil {
			return fmt.Errorf("invalid bool parameter value: %s", value)
		}
		fieldValue.SetBool(boolVal)
		return nil
	case reflect.Float64:
		var floatVal float64
		if floatVal, err = strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("invalid float parameter value: %s", value)
		}
		fieldValue.SetFloat(floatVal)
		return nil
	default:
		return fmt.Errorf("unsupported field type: %v", fieldValue.Kind())
	}

	return err
}

// setDefaultValue sets the default value from struct tag if available
func setDefaultValue(fieldValue reflect.Value, field reflect.StructField) error {
	defaultVal := field.Tag.Get("default")
	if defaultVal == "" {
		return nil
	}
	return setFieldValue(fieldValue, defaultVal)
}
