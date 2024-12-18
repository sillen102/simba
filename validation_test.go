package simba_test

import (
	"testing"

	"github.com/sillen102/simba"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Name     string `validate:"required"`
	Age      int    `validate:"gte=0,lte=130"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
}

func TestValidateStruct(t *testing.T) {
	tests := []struct {
		name          string
		input         interface{}
		paramType     simba.ParameterType
		expectedError bool
		errorCount    int
	}{
		{
			name: "Valid struct",
			input: TestStruct{
				Name:     "John Doe",
				Age:      25,
				Email:    "john@example.com",
				Password: "password123",
			},
			paramType:     simba.ParameterTypeBody,
			expectedError: false,
		},
		{
			name: "Invalid struct - missing required fields",
			input: TestStruct{
				Age: 25,
			},
			paramType:     simba.ParameterTypeBody,
			expectedError: true,
			errorCount:    3, // Name, Email, and Password are required
		},
		{
			name: "Invalid struct - invalid email",
			input: TestStruct{
				Name:     "John Doe",
				Age:      25,
				Email:    "invalid-email",
				Password: "password123",
			},
			paramType:     simba.ParameterTypeQuery,
			expectedError: true,
			errorCount:    1,
		},
		{
			name: "Invalid struct - age out of range",
			input: TestStruct{
				Name:     "John Doe",
				Age:      150,
				Email:    "john@example.com",
				Password: "password123",
			},
			paramType:     simba.ParameterTypePath,
			expectedError: true,
			errorCount:    1,
		},
		{
			name:          "Nil input",
			input:         nil,
			expectedError: false,
			errorCount:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := simba.ValidateStruct(tt.input, tt.paramType)

			if tt.expectedError {
				assert.NotNil(t, errors)
				assert.Equal(t, tt.errorCount, len(errors))

				// Check that each error has the correct parameter type
				for _, err := range errors {
					assert.Equal(t, tt.paramType, err.Type)
					assert.NotEmpty(t, err.Parameter)
					assert.NotEmpty(t, err.Message)
				}
			} else {
				assert.Nil(t, errors)
			}
		})
	}
}
