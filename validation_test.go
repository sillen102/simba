package simba_test

import (
	"testing"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/simbaTest/assert"
)

type TestStruct struct {
	Name     string            `validate:"required"`
	Age      int               `validate:"gte=0,lte=130"`
	Email    string            `validate:"required,email"`
	Password string            `validate:"required,min=8,max=20"`
	Tags     []string          `validate:"min=2,max=4"`
	Meta     map[string]string `validate:"min=1,max=3"`
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
				Tags:     []string{"go", "dev"},
				Meta:     map[string]string{"a": "1"},
			},
			expectedError: false,
		},
		{
			name: "Invalid struct - missing required fields",
			input: TestStruct{
				Age:  25,
				Tags: []string{"go", "dev"},
				Meta: map[string]string{"a": "1"},
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
				Tags:     []string{"go", "dev"},
				Meta:     map[string]string{"a": "1"},
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
				Tags:     []string{"go", "dev"},
				Meta:     map[string]string{"a": "1"},
			},
			expectedError: true,
			errorCount:    1,
			expectedMsgs:  []string{"age must be less than or equal to 130"},
		},
		// The rest of the cases (password/tags/meta validation) remain unchanged
		{
			name: "Password too short (min)",
			input: TestStruct{
				Name:     "John Doe",
				Age:      25,
				Email:    "john@example.com",
				Password: "short",
				Tags:     []string{"go", "dev"},
				Meta:     map[string]string{"a": "1"},
			},
			expectedError: true,
			errorCount:    1,
			expectedMsgs:  []string{"password must be at least 8 characters long"},
		},
		{
			name: "Password too long (max)",
			input: TestStruct{
				Name:     "John Doe",
				Age:      25,
				Email:    "john@example.com",
				Password: "thisisaverylongpasswordthatexceedsthemax",
				Tags:     []string{"go", "dev"},
				Meta:     map[string]string{"a": "1"},
			},
			expectedError: true,
			errorCount:    1,
			expectedMsgs:  []string{"password must not exceed 20 characters"},
		},
		{
			name: "Tags too few (min)",
			input: TestStruct{
				Name:     "John Doe",
				Age:      25,
				Email:    "john@example.com",
				Password: "password123",
				Tags:     []string{"go"},
				Meta:     map[string]string{"a": "1"},
			},
			expectedError: true,
			errorCount:    1,
			expectedMsgs:  []string{"tags must contain at least 2 items"},
		},
		{
			name: "Tags too many (max)",
			input: TestStruct{
				Name:     "John Doe",
				Age:      25,
				Email:    "john@example.com",
				Password: "password123",
				Tags:     []string{"go", "dev", "test", "extra", "overflow"},
				Meta:     map[string]string{"a": "1"},
			},
			expectedError: true,
			errorCount:    1,
			expectedMsgs:  []string{"tags must not contain more than 4 items"},
		},
		{
			name: "Meta too few (min)",
			input: TestStruct{
				Name:     "John Doe",
				Age:      25,
				Email:    "john@example.com",
				Password: "password123",
				Tags:     []string{"go", "dev"},
				Meta:     map[string]string{},
			},
			expectedError: true,
			errorCount:    1,
			expectedMsgs:  []string{"meta must contain at least 1 items"},
		},
		{
			name: "Meta too many (max)",
			input: TestStruct{
				Name:     "John Doe",
				Age:      25,
				Email:    "john@example.com",
				Password: "password123",
				Tags:     []string{"go", "dev"},
				Meta:     map[string]string{"a": "1", "b": "2", "c": "3", "d": "4"},
			},
			expectedError: true,
			errorCount:    1,
			expectedMsgs:  []string{"meta must not contain more than 3 items"},
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

				for _, err := range errors {
					assert.ContainsAnyOf(t, tt.expectedMsgs, err)
				}
			} else {
				assert.Nil(t, errors)
			}
		})
	}
}
