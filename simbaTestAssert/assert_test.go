package simbaTestAssert_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/sillen102/simba/simbaTestAssert"
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

func TestEqual(t *testing.T) {
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
				result = simbaTestAssert.Equal(mock, tc.expected, tc.actual, tc.message)
			} else {
				result = simbaTestAssert.Equal(mock, tc.expected, tc.actual)
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
			result := simbaTestAssert.Equal(mock, tc.expected, tc.actual)

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
				result = simbaTestAssert.NotEqual(mock, tc.expected, tc.actual, tc.message)
			} else {
				result = simbaTestAssert.NotEqual(mock, tc.expected, tc.actual)
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
			result := simbaTestAssert.NotEqual(mock, tc.expected, tc.actual)

			if result != tc.shouldPass {
				t.Errorf("NotEqual(%v, %v) returned %v, expected %v", tc.expected, tc.actual, result, tc.shouldPass)
			}
			if mock.failed == tc.shouldPass {
				t.Errorf("mockT.failed = %v, expected %v", mock.failed, !tc.shouldPass)
			}
		})
	}
}

func TestNilError(t *testing.T) {
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
				result = simbaTestAssert.NoError(mock, tc.err, tc.message)
			} else {
				result = simbaTestAssert.NoError(mock, tc.err)
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

func TestAssert(t *testing.T) {
	// Test passing assertion
	mock := &mockT{}
	result := simbaTestAssert.Assert(mock, true)
	if !result || mock.failed {
		t.Error("Assert should pass for true condition")
	}

	// Test failing assertion
	mock = &mockT{}
	result = simbaTestAssert.Assert(mock, false)
	if result || !mock.failed {
		t.Error("Assert should fail for false condition")
	}

	// Test with custom message
	mock = &mockT{}
	customMsg := "custom error message"
	simbaTestAssert.Assert(mock, false, customMsg)
	if !mock.failed || mock.errorMsg != customMsg {
		t.Errorf("Expected error message '%s', got '%s'", customMsg, mock.errorMsg)
	}

	// Test with formatted message
	mock = &mockT{}
	simbaTestAssert.Assert(mock, false, "failed with code %d", 404)
	expectedMsg := "failed with code 404"
	if !mock.failed || mock.errorMsg != expectedMsg {
		t.Errorf("Expected formatted error message '%s', got '%s'", expectedMsg, mock.errorMsg)
	}
}

func TestNil(t *testing.T) {
	// Test passing case with nil interface
	mock := &mockT{}
	var nilInterface interface{} = nil
	result := simbaTestAssert.Nil(mock, nilInterface)
	if !result || mock.failed {
		t.Error("Nil should pass for nil interface value")
	}

	// Test passing case with nil pointer
	mock = &mockT{}
	var nilPointer *string = nil
	result = simbaTestAssert.Nil(mock, nilPointer)
	if !result || mock.failed {
		t.Error("Nil should pass for nil pointer")
	}

	// Test failing case with non-nil value
	mock = &mockT{}
	nonNilValue := "test"
	result = simbaTestAssert.Nil(mock, nonNilValue)
	if result || !mock.failed {
		t.Error("Nil should fail for non-nil value")
	}

	// Test with custom message
	mock = &mockT{}
	customMsg := "custom error message"
	simbaTestAssert.Nil(mock, "not nil", customMsg)
	if !mock.failed || !contains(mock.errorMsg, customMsg) {
		t.Errorf("Expected error message to contain '%s', got '%s'", customMsg, mock.errorMsg)
	}

	// Test with formatted message
	mock = &mockT{}
	simbaTestAssert.Nil(mock, "not nil", "failed with code %d", 404)
	expectedMsg := "failed with code 404"
	if !mock.failed || !contains(mock.errorMsg, expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedMsg, mock.errorMsg)
	}
}

func TestNotNil(t *testing.T) {
	// Test passing case with non-nil value
	mock := &mockT{}
	nonNilValue := "test"
	result := simbaTestAssert.NotNil(mock, nonNilValue)
	if !result || mock.failed {
		t.Error("NotNil should pass for non-nil value")
	}

	// Test passing case with non-nil pointer
	mock = &mockT{}
	value := "test"
	nonNilPointer := &value
	result = simbaTestAssert.NotNil(mock, nonNilPointer)
	if !result || mock.failed {
		t.Error("NotNil should pass for non-nil pointer")
	}

	// Test failing case with nil interface
	mock = &mockT{}
	var nilInterface interface{} = nil
	result = simbaTestAssert.NotNil(mock, nilInterface)
	if result || !mock.failed {
		t.Error("NotNil should fail for nil interface")
	}

	// Test failing case with nil pointer
	mock = &mockT{}
	var nilPointer *string = nil
	result = simbaTestAssert.NotNil(mock, nilPointer)
	if result || !mock.failed {
		t.Error("NotNil should fail for nil pointer")
	}

	// Test with custom message
	mock = &mockT{}
	customMsg := "custom error message"
	simbaTestAssert.NotNil(mock, nil, customMsg)
	if !mock.failed || !contains(mock.errorMsg, customMsg) {
		t.Errorf("Expected error message to contain '%s', got '%s'", customMsg, mock.errorMsg)
	}

	// Test with formatted message
	mock = &mockT{}
	simbaTestAssert.NotNil(mock, nil, "failed with code %d", 404)
	expectedMsg := "failed with code 404"
	if !mock.failed || !contains(mock.errorMsg, expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedMsg, mock.errorMsg)
	}
}

// Helper function to check if a string contains another string
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
