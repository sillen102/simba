package simbaTestAssert_test

import (
	"testing"

	"github.com/sillen102/simba/simbaTestAssert"
)

func TestContainsOnly(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		expected   []int
		actual     []int
		message    string
		shouldPass bool
	}{
		{"identical slices", []int{1, 2, 3}, []int{1, 2, 3}, "", true},
		{"different order", []int{1, 2, 3}, []int{1, 3, 2}, "", false},
		{"different length", []int{1, 2}, []int{1, 2, 3}, "lengths should match", false},
		{"different elements", []int{1, 2, 3}, []int{1, 2, 4}, "", false},
		{"empty slices", []int{}, []int{}, "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockT{}
			var result bool
			if tc.message != "" {
				result = simbaTestAssert.ContainsOnly(mock, tc.expected, tc.actual, tc.message)
			} else {
				result = simbaTestAssert.ContainsOnly(mock, tc.expected, tc.actual)
			}

			if result != tc.shouldPass {
				t.Errorf("ContainsOnly(%v, %v) returned %v, expected %v", tc.expected, tc.actual, result, tc.shouldPass)
			}
			if mock.failed == tc.shouldPass {
				t.Errorf("mockT.failed = %v, expected %v", mock.failed, !tc.shouldPass)
			}
		})
	}
}

func TestContainsOnlyWithComplexStructs(t *testing.T) {
	t.Parallel()

	type Person struct {
		Name string
		Age  int
		Tags []string
	}

	alice := Person{Name: "Alice", Age: 30, Tags: []string{"dev", "manager"}}
	bob := Person{Name: "Bob", Age: 25, Tags: []string{"dev"}}
	charlie := Person{Name: "Charlie", Age: 35, Tags: []string{"qa"}}
	aliceClone := Person{Name: "Alice", Age: 30, Tags: []string{"dev", "manager"}}
	aliceModified := Person{Name: "Alice", Age: 30, Tags: []string{"dev", "lead"}}

	testCases := []struct {
		name       string
		expected   []Person
		actual     []Person
		shouldPass bool
	}{
		{
			name:       "identical complex structs",
			expected:   []Person{alice, bob},
			actual:     []Person{alice, bob},
			shouldPass: true,
		},
		{
			name:       "with clone equivalent",
			expected:   []Person{alice, bob},
			actual:     []Person{aliceClone, bob},
			shouldPass: true,
		},
		{
			name:       "modified nested slice",
			expected:   []Person{alice, bob},
			actual:     []Person{aliceModified, bob},
			shouldPass: false,
		},
		{
			name:       "different order",
			expected:   []Person{alice, bob},
			actual:     []Person{bob, alice},
			shouldPass: false,
		},
		{
			name:       "different length",
			expected:   []Person{alice, bob},
			actual:     []Person{alice, bob, charlie},
			shouldPass: false,
		},
		{
			name:       "empty slices",
			expected:   []Person{},
			actual:     []Person{},
			shouldPass: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockT{}
			result := simbaTestAssert.ContainsOnly(mock, tc.expected, tc.actual)

			if result != tc.shouldPass {
				t.Errorf("ContainsOnly test '%s' returned %v, expected %v",
					tc.name, result, tc.shouldPass)
			}
			if mock.failed == tc.shouldPass {
				t.Errorf("mockT.failed = %v, expected %v", mock.failed, !tc.shouldPass)
			}
		})
	}
}

func TestContainsOnlyInAnyOrder(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		expected   []int
		actual     []int
		message    string
		shouldPass bool
	}{
		{"identical slices", []int{1, 2, 3}, []int{1, 2, 3}, "", true},
		{"different order", []int{1, 2, 3}, []int{3, 1, 2}, "order shouldn't matter", true},
		{"different length", []int{1, 2}, []int{1, 2, 3}, "", false},
		{"different elements", []int{1, 2, 3}, []int{1, 2, 4}, "", false},
		{"duplicates in expected", []int{1, 2, 2}, []int{1, 2, 2}, "", true},
		{"duplicates in wrong count", []int{1, 2, 2}, []int{1, 2, 1}, "", false},
		{"empty slices", []int{}, []int{}, "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockT{}
			var result bool
			if tc.message != "" {
				result = simbaTestAssert.ContainsOnlyInAnyOrder(mock, tc.expected, tc.actual, tc.message)
			} else {
				result = simbaTestAssert.ContainsOnlyInAnyOrder(mock, tc.expected, tc.actual)
			}

			if result != tc.shouldPass {
				t.Errorf("ContainsOnlyInAnyOrder(%v, %v) returned %v, expected %v", tc.expected, tc.actual, result, tc.shouldPass)
			}
			if mock.failed == tc.shouldPass {
				t.Errorf("mockT.failed = %v, expected %v", mock.failed, !tc.shouldPass)
			}
		})
	}
}

func TestContainsOnlyInAnyOrderWithComplexStructs(t *testing.T) {
	t.Parallel()

	type Person struct {
		Name string
		Age  int
		Tags []string
	}

	alice := Person{Name: "Alice", Age: 30, Tags: []string{"dev", "manager"}}
	bob := Person{Name: "Bob", Age: 25, Tags: []string{"dev"}}
	charlie := Person{Name: "Charlie", Age: 35, Tags: []string{"qa"}}
	aliceClone := Person{Name: "Alice", Age: 30, Tags: []string{"dev", "manager"}}

	testCases := []struct {
		name       string
		expected   []Person
		actual     []Person
		shouldPass bool
	}{
		{
			name:       "identical complex structs",
			expected:   []Person{alice, bob},
			actual:     []Person{alice, bob},
			shouldPass: true,
		},
		{
			name:       "reordered complex structs",
			expected:   []Person{alice, bob},
			actual:     []Person{bob, alice},
			shouldPass: true,
		},
		{
			name:       "with clone of complex struct",
			expected:   []Person{alice, bob},
			actual:     []Person{bob, aliceClone},
			shouldPass: true,
		},
		{
			name:       "missing element",
			expected:   []Person{alice, bob, charlie},
			actual:     []Person{alice, bob},
			shouldPass: false,
		},
		{
			name:       "different length",
			expected:   []Person{alice, bob},
			actual:     []Person{alice, bob, charlie},
			shouldPass: false,
		},
		{
			name:       "with duplicates",
			expected:   []Person{alice, alice, bob},
			actual:     []Person{bob, alice, alice},
			shouldPass: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockT{}
			result := simbaTestAssert.ContainsOnlyInAnyOrder(mock, tc.expected, tc.actual)

			if result != tc.shouldPass {
				t.Errorf("ContainsOnlyInAnyOrder test '%s' returned %v, expected %v",
					tc.name, result, tc.shouldPass)
			}
			if mock.failed == tc.shouldPass {
				t.Errorf("mockT.failed = %v, expected %v", mock.failed, !tc.shouldPass)
			}
		})
	}
}

func TestContains(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		expected   any
		actual     any
		message    string
		shouldPass bool
	}{
		{"identical slices", []int{1, 2, 3}, []int{1, 2, 3}, "", true},
		{"single element in slice", 1, []int{1, 2, 3}, "", true},
		{"elements in string", []string{"a", "b", "c"}, "abc", "", true},
		{"sequence at beginning", []int{1, 2}, []int{1, 2, 3, 4}, "", true},
		{"sequence in middle", []int{2, 3}, []int{1, 2, 3, 4}, "", true},
		{"sequence at end", []int{3, 4}, []int{1, 2, 3, 4}, "", true},
		{"discontinuous sequence", []int{1, 3}, []int{1, 2, 3, 4}, "", false},
		{"out of order", []int{2, 1}, []int{1, 2, 3}, "", false},
		{"actual too short", []int{1, 2, 3}, []int{1, 2}, "too short", false},
		{"empty expected", []int{}, []int{1, 2, 3}, "", true},
		{"both empty", []int{}, []int{}, "", true},
		{"string contains substring", "ab", "abc", "", true},
		{"substring not contained", "abc", "ab", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockT{}
			var result bool
			if tc.message != "" {
				result = simbaTestAssert.Contains(mock, tc.expected, tc.actual, tc.message)
			} else {
				result = simbaTestAssert.Contains(mock, tc.expected, tc.actual)
			}

			if result != tc.shouldPass {
				t.Errorf("Contains(%v, %v) returned %v, expected %v", tc.expected, tc.actual, result, tc.shouldPass)
			}
			if mock.failed == tc.shouldPass {
				t.Errorf("mockT.failed = %v, expected %v", mock.failed, !tc.shouldPass)
			}
		})
	}
}

func TestContainsWithComplexStructs(t *testing.T) {
	t.Parallel()

	type Person struct {
		Name string
		Age  int
		Tags []string
	}

	alice := Person{Name: "Alice", Age: 30, Tags: []string{"dev", "manager"}}
	bob := Person{Name: "Bob", Age: 25, Tags: []string{"dev"}}
	charlie := Person{Name: "Charlie", Age: 35, Tags: []string{"qa"}}
	dave := Person{Name: "Dave", Age: 28, Tags: []string{"designer"}}
	aliceClone := Person{Name: "Alice", Age: 30, Tags: []string{"dev", "manager"}}

	testCases := []struct {
		name       string
		expected   any
		actual     []Person
		shouldPass bool
	}{
		{
			name:       "identical complex structs",
			expected:   []Person{alice, bob},
			actual:     []Person{alice, bob},
			shouldPass: true,
		},
		{
			name:       "single element in slice",
			expected:   bob,
			actual:     []Person{alice, bob, charlie},
			shouldPass: true,
		},
		{
			name:       "sequence at beginning",
			expected:   []Person{alice, bob},
			actual:     []Person{alice, bob, charlie},
			shouldPass: true,
		},
		{
			name:       "sequence in middle",
			expected:   []Person{bob, charlie},
			actual:     []Person{alice, bob, charlie, dave},
			shouldPass: true,
		},
		{
			name:       "sequence at end",
			expected:   []Person{charlie, dave},
			actual:     []Person{alice, bob, charlie, dave},
			shouldPass: true,
		},
		{
			name:       "with cloned struct",
			expected:   []Person{aliceClone, bob},
			actual:     []Person{alice, bob, charlie},
			shouldPass: true,
		},
		{
			name:       "out of order",
			expected:   []Person{bob, alice},
			actual:     []Person{alice, bob, charlie},
			shouldPass: false,
		},
		{
			name:       "discontinuous sequence",
			expected:   []Person{alice, charlie},
			actual:     []Person{alice, bob, charlie, dave},
			shouldPass: false,
		},
		{
			name:       "actual too short",
			expected:   []Person{alice, bob, charlie},
			actual:     []Person{alice, bob},
			shouldPass: false,
		},
		{
			name:       "empty expected",
			expected:   []Person{},
			actual:     []Person{alice, bob},
			shouldPass: true,
		},
		{
			name:       "both empty",
			expected:   []Person{},
			actual:     []Person{},
			shouldPass: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockT{}
			result := simbaTestAssert.Contains(mock, tc.expected, tc.actual)

			if result != tc.shouldPass {
				t.Errorf("Contains test '%s' returned %v, expected %v",
					tc.name, result, tc.shouldPass)
			}
			if mock.failed == tc.shouldPass {
				t.Errorf("mockT.failed = %v, expected %v", mock.failed, !tc.shouldPass)
			}
		})
	}
}

func TestContainsInAnyOrder(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		expected   []int
		actual     []int
		message    string
		shouldPass bool
	}{
		{"identical slices", []int{1, 2, 3}, []int{1, 2, 3}, "", true},
		{"different order", []int{3, 1, 2}, []int{1, 2, 3}, "", true},
		{"subset at beginning", []int{1, 2}, []int{1, 2, 3, 4}, "", true},
		{"subset scattered", []int{1, 3}, []int{1, 2, 3, 4}, "", true},
		{"with duplicates", []int{1, 1, 2}, []int{1, 1, 2, 3}, "", true},
		{"insufficient duplicates", []int{1, 1, 2}, []int{1, 2, 3}, "", false},
		{"missing element", []int{1, 5}, []int{1, 2, 3, 4}, "missing 5", false},
		{"actual too short", []int{1, 2, 3}, []int{1, 2}, "", false},
		{"empty expected", []int{}, []int{1, 2, 3}, "", true},
		{"both empty", []int{}, []int{}, "", true},
		{"slice in map", []int{}, []int{}, "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockT{}
			var result bool
			if tc.message != "" {
				result = simbaTestAssert.ContainsInAnyOrder(mock, tc.expected, tc.actual, tc.message)
			} else {
				result = simbaTestAssert.ContainsInAnyOrder(mock, tc.expected, tc.actual)
			}

			if result != tc.shouldPass {
				t.Errorf("ContainsInAnyOrder(%v, %v) returned %v, expected %v", tc.expected, tc.actual, result, tc.shouldPass)
			}
			if mock.failed == tc.shouldPass {
				t.Errorf("mockT.failed = %v, expected %v", mock.failed, !tc.shouldPass)
			}
		})
	}
}

func TestContainsInAnyOrderWithComplexStructs(t *testing.T) {
	t.Parallel()

	type Person struct {
		Name string
		Age  int
		Tags []string
	}

	alice := Person{Name: "Alice", Age: 30, Tags: []string{"dev", "manager"}}
	bob := Person{Name: "Bob", Age: 25, Tags: []string{"dev"}}
	charlie := Person{Name: "Charlie", Age: 35, Tags: []string{"qa"}}
	aliceClone := Person{Name: "Alice", Age: 30, Tags: []string{"dev", "manager"}}

	testCases := []struct {
		name       string
		expected   []Person
		actual     []Person
		shouldPass bool
	}{
		{
			name:       "identical complex structs",
			expected:   []Person{alice, bob},
			actual:     []Person{alice, bob},
			shouldPass: true,
		},
		{
			name:       "subset with different order",
			expected:   []Person{bob, alice},
			actual:     []Person{alice, charlie, bob},
			shouldPass: true,
		},
		{
			name:       "with clone in actual",
			expected:   []Person{alice},
			actual:     []Person{bob, aliceClone, charlie},
			shouldPass: true,
		},
		{
			name:       "missing element",
			expected:   []Person{alice, bob, charlie},
			actual:     []Person{alice, bob},
			shouldPass: false,
		},
		{
			name:       "with duplicates in expected",
			expected:   []Person{alice, alice},
			actual:     []Person{alice, aliceClone, bob},
			shouldPass: true,
		},
		{
			name:       "with insufficient duplicates",
			expected:   []Person{alice, alice, alice},
			actual:     []Person{alice, aliceClone},
			shouldPass: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockT{}
			result := simbaTestAssert.ContainsInAnyOrder(mock, tc.expected, tc.actual)

			if result != tc.shouldPass {
				t.Errorf("ContainsInAnyOrder test '%s' returned %v, expected %v",
					tc.name, result, tc.shouldPass)
			}
			if mock.failed == tc.shouldPass {
				t.Errorf("mockT.failed = %v, expected %v", mock.failed, !tc.shouldPass)
			}
		})
	}
}

func TestContainsAnyOf(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		expected   any
		actual     any
		message    string
		shouldPass bool
	}{
		{"identical slices", []int{1, 2, 3}, []int{1, 2, 3}, "", true},
		{"identical strings", "abc", "abc", "", true},
		{"not matching strings", "abc", "efg", "", false},
		{"single element slice", []int{1}, []int{1, 2, 3}, "", true},
		{"single element", 1, []int{1, 2, 3}, "", true},
		{"any from slice in element", []int{1, 2, 3}, 1, "", true},
		{"none from slice in element", []int{1, 2, 3}, 4, "", false},
		{"any from string slice in string", []string{"a", "b", "c"}, "a", "", true},
		{"all from string slice in string", []string{"a", "b", "c"}, "abc", "", true},
		{"none from string slice in string", []string{"a", "b", "c"}, "e", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockT{}
			var result bool
			if tc.message != "" {
				result = simbaTestAssert.ContainsAnyOf(mock, tc.expected, tc.actual, tc.message)
			} else {
				result = simbaTestAssert.ContainsAnyOf(mock, tc.expected, tc.actual)
			}

			if result != tc.shouldPass {
				t.Errorf("ContainsAnyOf(%v, %v) returned %v, expected %v", tc.expected, tc.actual, result, tc.shouldPass)
			}
			if mock.failed == tc.shouldPass {
				t.Errorf("mockT.failed = %v, expected %v", mock.failed, !tc.shouldPass)
			}
		})
	}
}

func TestContainsAnyOfWithComplexStructs(t *testing.T) {
	t.Parallel()

	type Person struct {
		Name string
		Age  int
		Tags []string
	}

	alice := Person{Name: "Alice", Age: 30, Tags: []string{"dev", "manager"}}
	bob := Person{Name: "Bob", Age: 25, Tags: []string{"dev"}}
	charlie := Person{Name: "Charlie", Age: 35, Tags: []string{"qa"}}
	dave := Person{Name: "Dave", Age: 28, Tags: []string{"designer"}}
	aliceClone := Person{Name: "Alice", Age: 30, Tags: []string{"dev", "manager"}}

	testCases := []struct {
		name       string
		expected   any
		actual     any
		shouldPass bool
	}{
		{
			name:       "identical complex structs",
			expected:   []Person{alice, bob},
			actual:     []Person{alice, bob},
			shouldPass: true,
		},
		{
			name:       "single struct matches element in slice",
			expected:   alice,
			actual:     []Person{alice, bob, charlie},
			shouldPass: true,
		},
		{
			name:       "slice contains single struct",
			expected:   []Person{alice, bob},
			actual:     alice,
			shouldPass: true, // alice exists in the expected slice
		},
		{
			name:       "slice doesn't contain single struct",
			expected:   []Person{alice, bob},
			actual:     dave,
			shouldPass: false, // dave doesn't exist in the expected slice
		},
		{
			name:       "with cloned struct",
			expected:   []Person{aliceClone, bob},
			actual:     alice,
			shouldPass: true, // alice equals aliceClone
		},
		{
			name:       "struct field matches in any element",
			expected:   []string{"Alice", "Bob"},
			actual:     alice.Name,
			shouldPass: true, // "Alice" is in the expected slice
		},
		{
			name:       "no struct field matches",
			expected:   []string{"Eva", "Frank"},
			actual:     alice.Name,
			shouldPass: false, // "Alice" is not in the expected slice
		},
		{
			name:       "complex slice contains any value",
			expected:   []Person{alice, bob, charlie},
			actual:     bob,
			shouldPass: true,
		},
		{
			name:       "compare string with struct fields",
			expected:   "Alice",
			actual:     []Person{alice, bob},
			shouldPass: false, // "Alice" isn't a direct match for any struct
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockT{}
			result := simbaTestAssert.ContainsAnyOf(mock, tc.expected, tc.actual)

			if result != tc.shouldPass {
				t.Errorf("ContainsAnyOf test '%s' returned %v, expected %v",
					tc.name, result, tc.shouldPass)
			}
			if mock.failed == tc.shouldPass {
				t.Errorf("mockT.failed = %v, expected %v", mock.failed, !tc.shouldPass)
			}
		})
	}
}

func TestNotEmpty(t *testing.T) {
	t.Parallel()

	// Define a struct type for testing
	type TestStruct struct {
		Field1 string
		Field2 int
	}

	// Create a channel for testing
	ch := make(chan int, 1)
	ch <- 1

	// Create an empty channel
	emptyCh := make(chan int)

	testCases := []struct {
		name       string
		value      any
		message    string
		shouldPass bool
	}{
		// Passing cases
		{"non-empty slice", []int{1, 2, 3}, "", true},
		{"non-empty string", "hello", "", true},
		{"non-empty map", map[string]int{"a": 1}, "", true},
		{"non-empty struct", TestStruct{Field1: "test", Field2: 123}, "", true},
		{"non-zero int", 42, "", true},
		{"non-zero float", 3.14, "", true},
		{"non-empty channel with buffer", ch, "", true},
		{"boolean true", true, "", true},

		// Failing cases
		{"nil value", nil, "", false},
		{"nil slice", []int(nil), "expected non-empty slice", false},
		{"nil map", map[string]int(nil), "", false},
		{"nil channel", chan int(nil), "", false},
		{"empty slice", []int{}, "", false},
		{"empty string", "", "string should not be empty", false},
		{"empty map", map[string]int{}, "", false},
		{"empty channel", emptyCh, "", false},
		{"zero struct", TestStruct{}, "", false},
		{"zero int", 0, "", false},
		{"zero float", 0.0, "", false},
		{"boolean false", false, "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockT{}
			var result bool

			if tc.message != "" {
				result = simbaTestAssert.NotEmpty(mock, tc.value, tc.message)
			} else {
				result = simbaTestAssert.NotEmpty(mock, tc.value)
			}

			if result != tc.shouldPass {
				t.Errorf("NotEmpty(%v) returned %v, expected %v", tc.value, result, tc.shouldPass)
			}
			if mock.failed == tc.shouldPass {
				t.Errorf("mockT.failed = %v, expected %v", mock.failed, !tc.shouldPass)
			}

			// Check if custom message is included in the error message
			if !tc.shouldPass && tc.message != "" && !contains(mock.errorMsg, tc.message) {
				t.Errorf("Expected error message to contain '%s', got '%s'", tc.message, mock.errorMsg)
			}
		})
	}

	// Test with formatted message arguments
	t.Run("formatted message", func(t *testing.T) {
		mock := &mockT{}
		result := simbaTestAssert.NotEmpty(mock, "", "expected non-empty value in test %d", 123)

		if result {
			t.Error("Should have failed with empty string")
		}

		expectedMsg := "expected non-empty value in test 123"
		if !contains(mock.errorMsg, expectedMsg) {
			t.Errorf("Expected error message to contain '%s', got '%s'", expectedMsg, mock.errorMsg)
		}
	})

	// Test with pointer types
	t.Run("pointer types", func(t *testing.T) {
		// Non-nil pointer
		nonNilPtr := &TestStruct{Field1: "test"}
		mock := &mockT{}
		result := simbaTestAssert.NotEmpty(mock, nonNilPtr)
		if !result || mock.failed {
			t.Error("NotEmpty should pass for non-nil pointer")
		}

		// Nil pointer
		var nilPtr *TestStruct = nil
		mock = &mockT{}
		result = simbaTestAssert.NotEmpty(mock, nilPtr)
		if result || !mock.failed {
			t.Error("NotEmpty should fail for nil pointer")
		}
	})
}
