package simbaOpenapi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi31"
)

// GetReflector creates a new OpenAPI reflector with custom options
func GetReflector() (*openapi31.Reflector, error) {
	r := openapi31.NewReflector()
	r.DefaultOptions = append(r.DefaultOptions, jsonschema.InterceptProp(func(params jsonschema.InterceptPropParams) error {
		if !params.Processed {
			return nil
		}

		if v, ok := params.Field.Tag.Lookup("validate"); ok {
			if strings.Contains(v, "required") {
				setIsRequired(params)
			}

			if strings.Contains(v, "min") {
				err := setMinProperty(params, v)
				if err != nil {
					return err
				}
			}

			if strings.Contains(v, "max") {
				err := setMaxProperty(params, v)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}))
	return r, nil
}

func setIsRequired(params jsonschema.InterceptPropParams) {
	params.ParentSchema.Required = append(params.ParentSchema.Required, params.Name)
}

func setMinProperty(params jsonschema.InterceptPropParams, v string) error {
	val, err := getFloatPropertyValue(v, "min")
	if err != nil {
		return err
	}
	params.PropertySchema.Minimum = &val
	return nil
}

func setMaxProperty(params jsonschema.InterceptPropParams, v string) error {
	val, err := getFloatPropertyValue(v, "max")
	if err != nil {
		return err
	}
	params.PropertySchema.Maximum = &val
	return nil
}

func getFloatPropertyValue(v string, propertyName string) (float64, error) {
	parts := strings.Split(v, propertyName+"=")
	if len(parts) <= 1 {
		return 0.0, fmt.Errorf("property %s not found", propertyName)
	}

	valStr := parts[1]
	if commaIdx := strings.Index(valStr, ","); commaIdx != -1 {
		valStr = valStr[:commaIdx]
	}

	value, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return 0.0, err
	}

	return value, nil
}
