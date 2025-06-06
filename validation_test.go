package simba_test

import (
	"testing"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/simbaTest/assert"
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
		expectedError bool
		errorCount    int
		expectedMsgs  any
	}{
		{
			name: "Valid struct",
			input: TestStruct{
				Name:     "John Doe",
				Age:      25,
				Email:    "john@example.com",
				Password: "password123",
			},
			expectedError: false,
		},
		{
			name: "Invalid struct - missing required fields",
			input: TestStruct{
				Age: 25,
			},
			expectedError: true,
			errorCount:    3, // Name, Email, and Password are required
			expectedMsgs:  []any{"name is required", "email is required", "password is required"},
		},
		{
			name: "Invalid struct - invalid email",
			input: TestStruct{
				Name:     "John Doe",
				Age:      25,
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedError: true,
			errorCount:    1,
			expectedMsgs:  []string{"'invalid-email' is not a valid email address"},
		},
		{
			name: "Invalid struct - age out of range",
			input: TestStruct{
				Name:     "John Doe",
				Age:      150,
				Email:    "john@example.com",
				Password: "password123",
			},
			expectedError: true,
			errorCount:    1,
			expectedMsgs:  []string{"age must be less than or equal to 130"},
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
			errors := simba.ValidateStruct(tt.input)

			if tt.expectedError {
				assert.NotNil(t, errors)
				assert.Equal(t, tt.errorCount, len(errors))

				// Check that each error has the correct parameter type
				for _, err := range errors {
					assert.ContainsAnyOf(t, tt.expectedMsgs, err)
				}
			} else {
				assert.Nil(t, errors)
			}
		})
	}
}
