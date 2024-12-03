package simba

import (
	"net/http"
	"reflect"
	"strconv"

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

		// Get value from request
		value := getParamValue(r, field)

		// Check for default value if no value was provided
		if value == "" && field.Tag.Get("default") != "" {
			value = field.Tag.Get("default")
		}

		// Handle special case for UUID
		if field.Type.String() == "uuid.UUID" {
			if err := handleUUIDField(fieldValue, value); err != nil {
				return instance, NewHttpError(http.StatusBadRequest, "invalid UUID parameter value", err,
					ValidationError{Parameter: field.Name, Type: getParamType(field), Message: "invalid UUID parameter value: " + value})
			}
			continue
		}

		// Set field value
		if err := setFieldValue(fieldValue, value); err != nil {
			return instance, NewHttpError(http.StatusBadRequest, "invalid parameter value", err,
				ValidationError{Parameter: field.Name, Type: getParamType(field), Message: "invalid parameter value: " + value})
		}
	}

	// Validate required fields
	if err := ValidateStruct(instance); len(err) > 0 {
		return instance, NewHttpError(http.StatusBadRequest, "missing required parameters", nil, err...)
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
	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Int, reflect.Int64:
		var intVal int64
		if intVal, err = strconv.ParseInt(value, 10, 64); err == nil {
			fieldValue.SetInt(intVal)
		}
	case reflect.Bool:
		var boolVal bool
		if boolVal, err = strconv.ParseBool(value); err == nil {
			fieldValue.SetBool(boolVal)
		}
	case reflect.Float64:
		var floatVal float64
		if floatVal, err = strconv.ParseFloat(value, 64); err == nil {
			fieldValue.SetFloat(floatVal)
		}
	}
	return err
}

// handleUUIDField handles the special case of UUID field types
func handleUUIDField(fieldValue reflect.Value, value string) error {
	if value == "" {
		return nil
	}
	uuidVal, err := uuid.Parse(value)
	if err != nil {
		return err
	}
	fieldValue.Set(reflect.ValueOf(uuidVal))
	return nil
}
