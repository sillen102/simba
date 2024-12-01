package simba

import (
	"net/http"
	"reflect"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// parseAndValidateParams creates a new instance of the parameter struct,
// populates it using the MapParams interface method, and validates it.
func parseAndValidateParams[Params any](r *http.Request) (*Params, error) {
	// Create an instance of the parameter struct
	var instance Params
	t := reflect.TypeOf(&instance).Elem()
	v := reflect.ValueOf(&instance).Elem()

	// Get path parameters from request
	params := httprouter.ParamsFromContext(r.Context())

	// Extract parameters from struct tags and set values
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

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

		if value == "" {
			continue
		}

		// Set value based on field type
		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(value)
		case reflect.Int, reflect.Int64:
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				validationError := ValidationError{
					Field:   field.Name,
					Message: "invalid integer parameter value: " + value,
				}
				return nil, NewHttpError(http.StatusBadRequest, "invalid parameter value", err, validationError)
			}
			fieldValue.SetInt(intVal)
		case reflect.Bool:
			boolVal, err := strconv.ParseBool(value)
			if err != nil {
				validationError := ValidationError{
					Field:   field.Name,
					Message: "invalid boolean parameter value: " + value,
				}
				return nil, NewHttpError(http.StatusBadRequest, "invalid parameter value", err, validationError)
			}
			fieldValue.SetBool(boolVal)
		}
	}

	// Validate required fields
	if err := ValidateStruct(instance); len(err) > 0 {
		return nil, NewHttpError(http.StatusBadRequest, "missing required parameters", nil, err...)
	}

	return &instance, nil
}
