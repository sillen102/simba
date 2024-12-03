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
	// If instance is NoParams, return early
	if _, ok := any(instance).(NoParams); ok {
		return instance, nil
	}
	t := reflect.TypeOf(&instance).Elem()
	// If instance is an empty struct, return early
	if t.NumField() == 0 {
		return instance, nil
	}
	v := reflect.ValueOf(&instance).Elem()

	// Get path parameters from request
	params := httprouter.ParamsFromContext(r.Context())

	// Extract parameters from struct tags and set values
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		var value string
		// Get value based on tag type
		switch {
		case field.Tag.Get("header") != "":
			value = r.Header.Get(field.Tag.Get("header"))
		case field.Tag.Get("path") != "":
			if params != nil {
				value = params.ByName(field.Tag.Get("path"))
			}
		case field.Tag.Get("query") != "":
			value = r.URL.Query().Get(field.Tag.Get("query"))
		}

		// Check for default value if no value was provided
		if value == "" {
			if defaultVal := field.Tag.Get("default"); defaultVal != "" {
				value = defaultVal
			} else {
				continue
			}
		}

		// Special handling for UUID type
		if field.Type.String() == "uuid.UUID" {
			if value == "" {
				continue
			}
			uuidVal, err := uuid.Parse(value)
			if err != nil {
				validationError := ValidationError{
					Field:   field.Name,
					Message: "invalid UUID parameter value: " + value,
				}
				return instance, NewHttpError(http.StatusBadRequest, "invalid parameter value", err, validationError)
			}
			fieldValue.Set(reflect.ValueOf(uuidVal))
			continue
		}

		// Set value based on field type
		var err error
		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(value)
		case reflect.Int, reflect.Int64:
			var intVal int64
			intVal, err = strconv.ParseInt(value, 10, 64)
			if err == nil {
				fieldValue.SetInt(intVal)
			}
		case reflect.Bool:
			var boolVal bool
			boolVal, err = strconv.ParseBool(value)
			if err == nil {
				fieldValue.SetBool(boolVal)
			}
		case reflect.Float64:
			var floatVal float64
			floatVal, err = strconv.ParseFloat(value, 64)
			if err == nil {
				fieldValue.SetFloat(floatVal)
			}
		}

		if err != nil {
			validationError := ValidationError{
				Field:   field.Name,
				Message: "invalid parameter value: " + value,
			}
			return instance, NewHttpError(http.StatusBadRequest, "invalid parameter value", err, validationError)
		}
	}

	// Validate required fields
	if err := ValidateStruct(instance); len(err) > 0 {
		return instance, NewHttpError(http.StatusBadRequest, "missing required parameters", nil, err...)
	}

	return instance, nil
}
