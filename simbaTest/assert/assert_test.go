package assert_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/sillen102/simba/simbaTest/assert"
)

// mockT implements the testing interface needed for our assertion functions
type mockT struct {
	failed    bool
	errorMsg  string
	errorArgs []interface{}
}

func (m *mockT) Errorf(format string, args ...interface{}) {
	m.failed = true
	m.errorMsg = format
	m.errorArgs = args
}

func (m *mockT) Helper() {
	// No-op for mock
}

func TestTrue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		condition  bool
		message    string
		shouldPass bool
	}{
		{"true condition", true, "", true},
		{"false condition", false, "condition should be true", false},
		{"false condition with formatted message", false, "condition failed with code %d", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockT{}
			var result bool
			if tc.message != "" {
				result = assert.True(mock, tc.condition, tc.message)
			} else {
				result = assert.True(mock, tc.condition)
			}

			if result != tc.shouldPass {
				t.Errorf("True(%v) returned %v, expected %v", tc.condition, result, tc.shouldPass)
			}
			if mock.failed == tc.shouldPass {
				t.Errorf("mockT.failed = %v, expected %v", mock.failed, !tc.shouldPass)
			}
		})
	}
}

func TestFalse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		condition  bool
		message    string
		shouldPass bool
	}{
		{"true condition", true, "condition should be false", false},
		{"false condition", false, "", true},
		{"false condition with formatted message", false, "condition failed with code %d", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockT{}
			var result bool
			if tc.message != "" {
				result = assert.False(mock, tc.condition, tc.message)
			} else {
				result = assert.False(mock, tc.condition)
			}

			if result != tc.shouldPass {
				t.Errorf("False(%v) returned %v, expected %v", tc.condition, result, tc.shouldPass)
			}
			if mock.failed == tc.shouldPass {
				t.Errorf("mockT.failed = %v, expected %v", mock.failed, !tc.shouldPass)
			}
		})
	}
}

func TestEqual(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		expected   interface{}
		actual     interface{}
		message    string
		shouldPass bool
	}{
		{"equal strings", "test", "test", "", true},
		{"different strings", "test", "other", "strings should match", false},
		{"equal ints", 42, 42, "", true},
		{"different ints", 42, 43, "", false},
		{"equal structs", struct{ a int }{1}, struct{ a int }{1}, "", true},
		{"different structs", struct{ a int }{1}, struct{ a int }{2}, "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockT{}
			var result bool
			if tc.message != "" {
				result = assert.Equal(mock, tc.expected, tc.actual, tc.message)
			} else {
				result = assert.Equal(mock, tc.expected, tc.actual)
			}

			if result != tc.shouldPass {
				t.Errorf("Equal(%v, %v) returned %v, expected %v", tc.expected, tc.actual, result, tc.shouldPass)
			}
			if mock.failed == tc.shouldPass {
				t.Errorf("mockT.failed = %v, expected %v", mock.failed, !tc.shouldPass)
			}
		})
	}
}

func TestEqualWithNestedStructs(t *testing.T) {
	t.Parallel()

	type Inner struct {
		Value string
		Count int
	}

	type Outer struct {
		Name  string
		Inner Inner
		Data  map[string]interface{}
	}

	testCases := []struct {
		name       string
		expected   interface{}
		actual     interface{}
		shouldPass bool
	}{
		{
			name: "identical nested structs",
			expected: Outer{
				Name: "test",
				Inner: Inner{
					Value: "inner",
					Count: 42,
				},
				Data: map[string]interface{}{"key": "value"},
			},
			actual: Outer{
				Name: "test",
				Inner: Inner{
					Value: "inner",
					Count: 42,
				},
				Data: map[string]interface{}{"key": "value"},
			},
			shouldPass: true,
		},
		{
			name: "different inner struct field",
			expected: Outer{
				Name: "test",
				Inner: Inner{
					Value: "inner",
					Count: 42,
				},
			},
			actual: Outer{
				Name: "test",
				Inner: Inner{
					Value: "different",
					Count: 42,
				},
			},
			shouldPass: false,
		},
		{
			name: "different map content",
			expected: Outer{
				Name:  "test",
				Inner: Inner{Value: "inner", Count: 42},
				Data:  map[string]interface{}{"key": "value"},
			},
			actual: Outer{
				Name:  "test",
				Inner: Inner{Value: "inner", Count: 42},
				Data:  map[string]interface{}{"key": "different"},
			},
			shouldPass: false,
		},
		{
			name: "deeply nested structs",
			expected: map[string]map[string][]Inner{
				"layer1": {
					"layer2": {
						{Value: "test1", Count: 1},
						{Value: "test2", Count: 2},
					},
				},
			},
			actual: map[string]map[string][]Inner{
				"layer1": {
					"layer2": {
						{Value: "test1", Count: 1},
						{Value: "test2", Count: 2},
					},
				},
			},
			shouldPass: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockT{}
			result := assert.Equal(mock, tc.expected, tc.actual)

			if result != tc.shouldPass {
				t.Errorf("Equal(%v, %v) returned %v, expected %v", tc.expected, tc.actual, result, tc.shouldPass)
			}
			if mock.failed == tc.shouldPass {
				t.Errorf("mockT.failed = %v, expected %v", mock.failed, !tc.shouldPass)
			}
		})
	}
}

func TestNotEqual(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		expected   interface{}
		actual     interface{}
		message    string
		shouldPass bool
	}{
		{"equal strings", "test", "test", "", false},
		{"different strings", "test", "other", "strings should differ", true},
		{"equal ints", 42, 42, "", false},
		{"different ints", 42, 43, "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockT{}
			var result bool
			if tc.message != "" {
				result = assert.NotEqual(mock, tc.expected, tc.actual, tc.message)
			} else {
				result = assert.NotEqual(mock, tc.expected, tc.actual)
			}

			if result != tc.shouldPass {
				t.Errorf("NotEqual(%v, %v) returned %v, expected %v", tc.expected, tc.actual, result, tc.shouldPass)
			}
			if mock.failed == tc.shouldPass {
				t.Errorf("mockT.failed = %v, expected %v", mock.failed, !tc.shouldPass)
			}
		})
	}
}

func TestNotEqualWithNestedStructs(t *testing.T) {
	type Inner struct {
		Value string
		Count int
	}

	type Outer struct {
		Name  string
		Inner Inner
		Data  map[string]interface{}
	}

	testCases := []struct {
		name       string
		expected   interface{}
		actual     interface{}
		shouldPass bool
	}{
		{
			name: "identical nested structs",
			expected: Outer{
				Name: "test",
				Inner: Inner{
					Value: "inner",
					Count: 42,
				},
				Data: map[string]interface{}{"key": "value"},
			},
			actual: Outer{
				Name: "test",
				Inner: Inner{
					Value: "inner",
					Count: 42,
				},
				Data: map[string]interface{}{"key": "value"},
			},
			shouldPass: false, // NotEqual should fail for identical structs
		},
		{
			name: "different inner struct field",
			expected: Outer{
				Name: "test",
				Inner: Inner{
					Value: "inner",
					Count: 42,
				},
			},
			actual: Outer{
				Name: "test",
				Inner: Inner{
					Value: "different",
					Count: 42,
				},
			},
			shouldPass: true, // NotEqual should pass for different structs
		},
		{
			name: "deeply nested structs - different",
			expected: map[string]map[string][]Inner{
				"layer1": {
					"layer2": {
						{Value: "test1", Count: 1},
						{Value: "test2", Count: 2},
					},
				},
			},
			actual: map[string]map[string][]Inner{
				"layer1": {
					"layer2": {
						{Value: "test1", Count: 1},
						{Value: "test2", Count: 3}, // Different Count
					},
				},
			},
			shouldPass: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockT{}
			result := assert.NotEqual(mock, tc.expected, tc.actual)

			if result != tc.shouldPass {
				t.Errorf("NotEqual(%v, %v) returned %v, expected %v", tc.expected, tc.actual, result, tc.shouldPass)
			}
			if mock.failed == tc.shouldPass {
				t.Errorf("mockT.failed = %v, expected %v", mock.failed, !tc.shouldPass)
			}
		})
	}
}

func TestNoError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		err        error
		message    string
		shouldPass bool
	}{
		{"nil error", nil, "", true},
		{"non-nil error", errors.New("test error"), "error should be nil", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockT{}
			var result bool
			if tc.message != "" {
				result = assert.NoError(mock, tc.err, tc.message)
			} else {
				result = assert.NoError(mock, tc.err)
			}

			if result != tc.shouldPass {
				t.Errorf("NilError(%v) returned %v, expected %v", tc.err, result, tc.shouldPass)
			}
			if mock.failed == tc.shouldPass {
				t.Errorf("mockT.failed = %v, expected %v", mock.failed, !tc.shouldPass)
			}
		})
	}
}

func TestError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		err        error
		message    string
		shouldPass bool
	}{
		{"nil error", nil, "error should not be nil", false},
		{"non-nil error", errors.New("test error"), "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockT{}
			var result bool
			if tc.message != "" {
				result = assert.Error(mock, tc.err, tc.message)
			} else {
				result = assert.Error(mock, tc.err)
			}

			if result != tc.shouldPass {
				t.Errorf("Error(%v) returned %v, expected %v", tc.err, result, tc.shouldPass)
			}
			if mock.failed == tc.shouldPass {
				t.Errorf("mockT.failed = %v, expected %v", mock.failed, !tc.shouldPass)
			}
		})
	}
}

func TestAssert(t *testing.T) {
	t.Parallel()

	// Test passing assertion
	mock := &mockT{}
	result := assert.Assert(mock, true)
	if !result || mock.failed {
		t.Error("Assert should pass for true condition")
	}

	// Test failing assertion
	mock = &mockT{}
	result = assert.Assert(mock, false)
	if result || !mock.failed {
		t.Error("Assert should fail for false condition")
	}

	// Test with custom message
	mock = &mockT{}
	customMsg := "custom error message"
	assert.Assert(mock, false, customMsg)
	if !mock.failed || mock.errorMsg != customMsg {
		t.Errorf("Expected error message '%s', got '%s'", customMsg, mock.errorMsg)
	}

	// Test with formatted message
	mock = &mockT{}
	assert.Assert(mock, false, "failed with code %d", 404)
	expectedMsg := "failed with code 404"
	if !mock.failed || mock.errorMsg != expectedMsg {
		t.Errorf("Expected formatted error message '%s', got '%s'", expectedMsg, mock.errorMsg)
	}
}

func TestNil(t *testing.T) {
	t.Parallel()

	// Test passing case with nil interface
	mock := &mockT{}
	var nilInterface interface{} = nil
	result := assert.Nil(mock, nilInterface)
	if !result || mock.failed {
		t.Error("Nil should pass for nil interface value")
	}

	// Test passing case with nil pointer
	mock = &mockT{}
	var nilPointer *string = nil
	result = assert.Nil(mock, nilPointer)
	if !result || mock.failed {
		t.Error("Nil should pass for nil pointer")
	}

	// Test failing case with non-nil value
	mock = &mockT{}
	nonNilValue := "test"
	result = assert.Nil(mock, nonNilValue)
	if result || !mock.failed {
		t.Error("Nil should fail for non-nil value")
	}

	// Test with custom message
	mock = &mockT{}
	customMsg := "custom error message"
	assert.Nil(mock, "not nil", customMsg)
	if !mock.failed || !contains(mock.errorMsg, customMsg) {
		t.Errorf("Expected error message to contain '%s', got '%s'", customMsg, mock.errorMsg)
	}

	// Test with formatted message
	mock = &mockT{}
	assert.Nil(mock, "not nil", "failed with code %d", 404)
	expectedMsg := "failed with code 404"
	if !mock.failed || !contains(mock.errorMsg, expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedMsg, mock.errorMsg)
	}
}

func TestNotNil(t *testing.T) {
	t.Parallel()

	// Test passing case with non-nil value
	mock := &mockT{}
	nonNilValue := "test"
	result := assert.NotNil(mock, nonNilValue)
	if !result || mock.failed {
		t.Error("NotNil should pass for non-nil value")
	}

	// Test passing case with non-nil pointer
	mock = &mockT{}
	value := "test"
	nonNilPointer := &value
	result = assert.NotNil(mock, nonNilPointer)
	if !result || mock.failed {
		t.Error("NotNil should pass for non-nil pointer")
	}

	// Test failing case with nil interface
	mock = &mockT{}
	var nilInterface interface{} = nil
	result = assert.NotNil(mock, nilInterface)
	if result || !mock.failed {
		t.Error("NotNil should fail for nil interface")
	}

	// Test failing case with nil pointer
	mock = &mockT{}
	var nilPointer *string = nil
	result = assert.NotNil(mock, nilPointer)
	if result || !mock.failed {
		t.Error("NotNil should fail for nil pointer")
	}

	// Test with custom message
	mock = &mockT{}
	customMsg := "custom error message"
	assert.NotNil(mock, nil, customMsg)
	if !mock.failed || !contains(mock.errorMsg, customMsg) {
		t.Errorf("Expected error message to contain '%s', got '%s'", customMsg, mock.errorMsg)
	}

	// Test with formatted message
	mock = &mockT{}
	assert.NotNil(mock, nil, "failed with code %d", 404)
	expectedMsg := "failed with code 404"
	if !mock.failed || !contains(mock.errorMsg, expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedMsg, mock.errorMsg)
	}
}

// Helper function to check if a string contains another string
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
