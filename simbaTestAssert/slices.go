package simbaTestAssert

import (
	"fmt"
	"reflect"
)

// ContainsOnly compares two slices and returns true if they contain exactly the same elements in the same order.
// If not, it formats an error message and reports it through the test interface.
func ContainsOnly[T any](t interface {
	Errorf(format string, args ...any)
}, expected, actual []T, msgAndArgs ...any) bool {
	if len(expected) != len(actual) {
		message := formatSliceFailureMessage("Slices have different lengths", expected, actual, msgAndArgs...)
		t.Errorf("%s", message)
		return false
	}

	for i := 0; i < len(expected); i++ {
		if !reflect.DeepEqual(expected[i], actual[i]) {
			message := formatSliceFailureMessage(fmt.Sprintf("Slices differ at index %d", i), expected, actual, msgAndArgs...)
			t.Errorf("%s", message)
			return false
		}
	}

	return true
}

// ContainsOnlyInAnyOrder compares two slices and returns true if they contain exactly the same elements in any order.
// If not, it formats an error message and reports it through the test interface.
func ContainsOnlyInAnyOrder[T any](t interface {
	Errorf(format string, args ...any)
}, expected, actual []T, msgAndArgs ...any) bool {
	if len(expected) != len(actual) {
		message := formatSliceFailureMessage("Slices have different lengths", expected, actual, msgAndArgs...)
		t.Errorf("%s", message)
		return false
	}

	// Create a "matching" array to track which elements have been matched
	matched := make([]bool, len(actual))

	// For each element in expected, find a match in actual
	for _, expectedItem := range expected {
		found := false
		for i, actualItem := range actual {
			if !matched[i] && reflect.DeepEqual(expectedItem, actualItem) {
				matched[i] = true
				found = true
				break
			}
		}

		if !found {
			message := formatSliceFailureMessage(
				fmt.Sprintf("Element '%v' in expected was not found in actual or appears too few times", expectedItem),
				expected, actual, msgAndArgs...)
			t.Errorf("%s", message)
			return false
		}
	}

	// All elements in expected have matches in actual, and lengths are equal,
	// so the slices contain exactly the same elements in any order
	return true
}

// Contains checks if the actual slice contains all elements from the expected slice in the same order.
// Extra elements in the actual slice are allowed.
func Contains[T any](t interface {
	Errorf(format string, args ...any)
}, expected, actual []T, msgAndArgs ...any) bool {
	if len(expected) == 0 {
		return true
	}

	if len(actual) < len(expected) {
		message := formatSliceFailureMessage("Actual slice is too short to contain expected elements", expected, actual, msgAndArgs...)
		t.Errorf("%s", message)
		return false
	}

	// Try to find the expected sequence starting at each possible position
	for startPos := 0; startPos <= len(actual)-len(expected); startPos++ {
		found := true
		for i := 0; i < len(expected); i++ {
			if !reflect.DeepEqual(expected[i], actual[startPos+i]) {
				found = false
				break
			}
		}
		if found {
			return true
		}
	}

	message := formatSliceFailureMessage("Expected sequence not found in actual slice", expected, actual, msgAndArgs...)
	t.Errorf("%s", message)
	return false
}

// ContainsInAnyOrder checks if the actual slice contains all elements from the expected slice in any order.
// Extra elements in the actual slice are allowed.
func ContainsInAnyOrder[T any](t interface {
	Errorf(format string, args ...any)
}, expected, actual []T, msgAndArgs ...any) bool {
	if len(expected) == 0 {
		return true
	}

	if len(actual) < len(expected) {
		message := formatSliceFailureMessage("Actual slice has fewer elements than expected", expected, actual, msgAndArgs...)
		t.Errorf("%s", message)
		return false
	}

	// For each element in expected, find and "remove" a match from actual
	// We'll use a copy of actual so we don't modify the original
	remainingActual := make([]T, len(actual))
	copy(remainingActual, actual)

	// Keep track of which items we've already matched
	matched := make([]bool, len(remainingActual))

	for _, expectedItem := range expected {
		found := false
		for i, actualItem := range remainingActual {
			if !matched[i] && reflect.DeepEqual(expectedItem, actualItem) {
				matched[i] = true
				found = true
				break
			}
		}

		if !found {
			message := formatSliceFailureMessage(
				fmt.Sprintf("Expected item '%v' not found in actual slice or appears too few times", expectedItem),
				expected, actual, msgAndArgs...)
			t.Errorf("%s", message)
			return false
		}
	}

	return true
}

// formatSliceFailureMessage creates a descriptive failure message for unequal slices
func formatSliceFailureMessage(reason string, expected, actual any, msgAndArgs ...any) string {
	var msg string
	if len(msgAndArgs) > 0 {
		if fmtMsg, ok := msgAndArgs[0].(string); ok {
			if len(msgAndArgs) > 1 {
				msg = fmt.Sprintf(fmtMsg, msgAndArgs[1:]...) + "\n"
			} else {
				msg = fmtMsg + "\n"
			}
		}
	}

	return fmt.Sprintf("%s%s\nExpected: %#v\nActual:   %#v",
		msg,
		reason,
		expected,
		actual)
}
