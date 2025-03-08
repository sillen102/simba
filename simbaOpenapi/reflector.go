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
	val, err := getPropertyValue[float64](v, "min")
	if err != nil {
		return err
	}
	params.ParentSchema.Minimum = &val
	return nil
}

func setMaxProperty(params jsonschema.InterceptPropParams, v string) error {
	val, err := getPropertyValue[float64](v, "max")
	if err != nil {
		return err
	}
	params.ParentSchema.Maximum = &val
	return nil
}

func getPropertyValue[T any](v string, propertyName string) (T, error) {
	var zero T

	parts := strings.Split(v, propertyName+"=")
	if len(parts) <= 1 {
		return zero, fmt.Errorf("property %s not found", propertyName)
	}

	valStr := parts[1]
	if commaIdx := strings.Index(valStr, ","); commaIdx != -1 {
		valStr = valStr[:commaIdx]
	}

	var result T
	switch any(zero).(type) {
	case int, int8, int16, int32, int64:
		v, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			return zero, err
		}
		result = any(v).(T)
	case float32, float64:
		v, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return zero, err
		}
		result = any(v).(T)
	case string:
		result = any(valStr).(T)
	case bool:
		v, err := strconv.ParseBool(valStr)
		if err != nil {
			return zero, err
		}
		result = any(v).(T)
	default:
		return zero, fmt.Errorf("unsupported type for property %s", propertyName)
	}

	return result, nil
}
