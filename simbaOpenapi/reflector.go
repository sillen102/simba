package simbaOpenapi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi31"
)

const MIN = "min"
const MAX = "max"

// GetReflector creates a new OpenAPI reflector with custom options.
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

			if strings.Contains(v, MIN) {
				err := setMinProperty(params, v)
				if err != nil {
					return err
				}
			}

			if strings.Contains(v, MAX) {
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
	switch {
	case hasSimpleType(params):
		switch *params.PropertySchema.Type.SimpleTypes {
		case jsonschema.String:
			val, err := parseTagInt(v, MIN)
			if err != nil {
				return err
			}
			params.PropertySchema.MinLength = val
			return nil
		case jsonschema.Array:
			val, err := parseTagInt(v, MIN)
			if err != nil {
				return err
			}
			params.PropertySchema.MinItems = val
			return nil
		case jsonschema.Number, jsonschema.Integer:
			val, err := parseTagFloat(v, MIN)
			if err != nil {
				return err
			}
			params.PropertySchema.Minimum = &val
			return nil
		case jsonschema.Boolean, jsonschema.Null, jsonschema.Object:
			return nil
		}
	case isSliceArrayType(params):
		val, err := parseTagInt(v, MIN)
		if err != nil {
			return err
		}
		params.PropertySchema.MinItems = val
		return nil
	default:
		val, err := parseTagFloat(v, MIN)
		if err != nil {
			return err
		}
		params.PropertySchema.Minimum = &val
		return nil
	}

	return nil
}

func setMaxProperty(params jsonschema.InterceptPropParams, v string) error {
	switch {
	case params.PropertySchema.Type != nil && params.PropertySchema.Type.SimpleTypes != nil:
		switch *params.PropertySchema.Type.SimpleTypes {
		case jsonschema.String:
			val, err := parseTagInt(v, MAX)
			if err != nil {
				return err
			}
			params.PropertySchema.MaxLength = &val
			return nil
		case jsonschema.Array:
			val, err := parseTagInt(v, MAX)
			if err != nil {
				return err
			}
			params.PropertySchema.MaxItems = &val
			return nil
		case jsonschema.Number, jsonschema.Integer:
			val, err := parseTagFloat(v, MAX)
			if err != nil {
				return err
			}
			params.PropertySchema.Maximum = &val
			return nil
		case jsonschema.Boolean, jsonschema.Null, jsonschema.Object:
			return nil
		}
	case isSliceArrayType(params):
		val, err := parseTagInt(v, MAX)
		if err != nil {
			return err
		}
		params.PropertySchema.MaxItems = &val
		return nil
	default:
		val, err := parseTagFloat(v, MAX)
		if err != nil {
			return err
		}
		params.PropertySchema.Maximum = &val
		return nil
	}

	return nil
}

func hasSimpleType(params jsonschema.InterceptPropParams) bool {
	return params.PropertySchema.Type != nil && params.PropertySchema.Type.SimpleTypes != nil
}

func isSliceArrayType(params jsonschema.InterceptPropParams) bool {
	return params.PropertySchema.Type != nil &&
		len(params.PropertySchema.Type.SliceOfSimpleTypeValues) > 0 &&
		params.PropertySchema.Type.SliceOfSimpleTypeValues[0] == jsonschema.Array
}

// parseTagInt extracts a named value from a validate tag string (e.g. "required,min=5,max=10")
// and parses it as int64. Used for count-based constraints (MinLength, MinItems, MaxLength, MaxItems).
func parseTagInt(v string, propertyName string) (int64, error) {
	parts := strings.Split(v, propertyName+"=")
	if len(parts) <= 1 {
		return 0, fmt.Errorf("property %s not found", propertyName)
	}

	valStr := parts[1]
	if commaIdx := strings.Index(valStr, ","); commaIdx != -1 {
		valStr = valStr[:commaIdx]
	}

	value, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return value, nil
}

// parseTagFloat extracts a named value from a validate tag string (e.g. "required,min=1.5,max=9.9")
// and parses it as float64. Used for value-based constraints (Minimum, Maximum) — JSON Schema defines
// these fields as number even for integer types.
func parseTagFloat(v string, propertyName string) (float64, error) {
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
