package simbaTestAssert_test

import (
	"errors"
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
				result = simbaTestAssert.NilError(mock, tc.err, tc.message)
			} else {
				result = simbaTestAssert.NilError(mock, tc.err)
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
