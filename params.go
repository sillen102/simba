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
			if params := httprouter.ParamsFromContext(r.Context()); params != nil {
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
				return nil, NewHttpError(http.StatusBadRequest, "invalid integer parameter", err)
			}
			fieldValue.SetInt(intVal)
		}
	}

	// Validate required fields
	if err := ValidateStruct(instance); len(err) > 0 {
		return nil, NewHttpError(http.StatusBadRequest, "missing required parameters", nil, err...)
	}

	return &instance, nil
}
