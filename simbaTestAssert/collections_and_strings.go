package simbaTestAssert

import (
	"fmt"
	"reflect"
	"strings"
)

// ContainsOnly compares two slices and returns true if they contain exactly the same elements in the same order.
// If not, it formats an error message and reports it through the test interface.
func ContainsOnly[T any](t interface {
	Errorf(format string, args ...any)
	Helper()
}, expected, actual []T, msgAndArgs ...any) bool {
	t.Helper()

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
	Helper()
}, expected, actual []T, msgAndArgs ...any) bool {
	t.Helper()

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

// Contains checks if an item is present in a collection, or if a collection contains another collection
func Contains(t interface {
	Errorf(format string, args ...any)
	Helper()
}, item any, collection any, msgAndArgs ...any) bool {
	t.Helper()

	if collection == nil {
		message := fmt.Sprintf("Cannot check if nil contains '%v'", item)
		if len(msgAndArgs) > 0 {
			if msgFormat, ok := msgAndArgs[0].(string); ok {
				customMsg := formatMessage(msgFormat, msgAndArgs[1:]...)
				message = fmt.Sprintf("%s (Cannot check if nil contains '%v')", customMsg, item)
			}
		}
		t.Errorf(message)
		return false
	}

	v := reflect.ValueOf(collection)

	switch v.Kind() {
	case reflect.String:
		// Handle string containing substring case
		str := collection.(string)

		// Check if item is a single string
		itemStr, ok := item.(string)
		if ok {
			if strings.Contains(str, itemStr) {
				return true
			}

			message := fmt.Sprintf("Expected '%s' to contain '%s'", str, itemStr)
			if len(msgAndArgs) > 0 {
				if msgFormat, ok := msgAndArgs[0].(string); ok {
					customMsg := formatMessage(msgFormat, msgAndArgs[1:]...)
					message = fmt.Sprintf("%s (Expected '%s' to contain '%s')", customMsg, str, itemStr)
				}
			}
			t.Errorf(message)
			return false
		}

		// Check if item is a slice of strings
		itemVal := reflect.ValueOf(item)
		if itemVal.Kind() == reflect.Slice || itemVal.Kind() == reflect.Array {
			allFound := true
			missing := []any{}

			for i := 0; i < itemVal.Len(); i++ {
				element := itemVal.Index(i).Interface()
				elemStr, ok := element.(string)
				if !ok {
					t.Errorf("For string contains check, slice elements must be strings")
					return false
				}

				if !strings.Contains(str, elemStr) {
					allFound = false
					missing = append(missing, elemStr)
				}
			}

			if allFound {
				return true
			}

			message := fmt.Sprintf("Expected '%s' to contain elements %v, missing %v", str, item, missing)
			if len(msgAndArgs) > 0 {
				if msgFormat, ok := msgAndArgs[0].(string); ok {
					customMsg := formatMessage(msgFormat, msgAndArgs[1:]...)
					message = fmt.Sprintf("%s (%s)", customMsg, message)
				}
			}
			t.Errorf(message)
			return false
		}

		t.Errorf("For string contains check, item must be a string or slice of strings")
		return false

	case reflect.Slice, reflect.Array:
		// Check if item is a single element in the collection
		itemVal := reflect.ValueOf(item)
		if itemVal.Kind() != reflect.Slice && itemVal.Kind() != reflect.Array {
			// Search for a single item in the collection
			length := v.Len()
			for i := 0; i < length; i++ {
				if reflect.DeepEqual(v.Index(i).Interface(), item) {
					return true
				}
			}
			message := fmt.Sprintf("Expected %v to contain '%v'", collection, item)
			if len(msgAndArgs) > 0 {
				if msgFormat, ok := msgAndArgs[0].(string); ok {
					customMsg := formatMessage(msgFormat, msgAndArgs[1:]...)
					message = fmt.Sprintf("%s (Expected %v to contain '%v')", customMsg, collection, item)
				}
			}
			t.Errorf(message)
			return false
		} else {
			// Check if the collection contains a sequence (item is a slice/array)
			itemLen := itemVal.Len()
			if itemLen == 0 {
				return true // Empty sequence is always contained
			}

			collectionLen := v.Len()
			if itemLen > collectionLen {
				message := fmt.Sprintf("Expected %v to contain %v, but collection is shorter", collection, item)
				if len(msgAndArgs) > 0 {
					if msgFormat, ok := msgAndArgs[0].(string); ok {
						customMsg := formatMessage(msgFormat, msgAndArgs[1:]...)
						message = fmt.Sprintf("%s (Expected %v to contain %v, but collection is shorter)", customMsg, collection, item)
					}
				}
				t.Errorf(message)
				return false
			}

			// Check for consecutive sequence
			for i := 0; i <= collectionLen-itemLen; i++ {
				match := true
				for j := 0; j < itemLen; j++ {
					if !reflect.DeepEqual(v.Index(i+j).Interface(), itemVal.Index(j).Interface()) {
						match = false
						break
					}
				}
				if match {
					return true
				}
			}

			message := fmt.Sprintf("Expected %v to contain sequence %v", collection, item)
			if len(msgAndArgs) > 0 {
				if msgFormat, ok := msgAndArgs[0].(string); ok {
					customMsg := formatMessage(msgFormat, msgAndArgs[1:]...)
					message = fmt.Sprintf("%s (Expected %v to contain sequence %v)", customMsg, collection, item)
				}
			}
			t.Errorf(message)
			return false
		}

	default:
		t.Errorf("Cannot check contains on type %T", collection)
		return false
	}
}

// ContainsInAnyOrder checks if the actual slice contains all elements from the expected slice in any order.
// Extra elements in the actual slice are allowed.
func ContainsInAnyOrder[T any](t interface {
	Errorf(format string, args ...any)
	Helper()
}, expected, actual []T, msgAndArgs ...any) bool {
	t.Helper()

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

// ContainsAnyOf checks if the collection contains any elements from expected.
// For slice expected vs single value collection: checks if any element from the slice equals the single value.
// For string expected vs slice of strings collection: checks if each character in expected exists in any element of collection.
// For slice of strings expected vs string collection: checks if any string from the slice is contained in the collection string.
func ContainsAnyOf(t interface {
	Errorf(format string, args ...any)
	Helper()
}, expected any, collection any, msgAndArgs ...any) bool {
	t.Helper()

	if collection == nil {
		message := fmt.Sprintf("Cannot check if nil contains any elements from '%v'", expected)
		if len(msgAndArgs) > 0 {
			if msgFormat, ok := msgAndArgs[0].(string); ok {
				customMsg := formatMessage(msgFormat, msgAndArgs[1:]...)
				message = fmt.Sprintf("%s (Cannot check if nil contains any elements from '%v')", customMsg, expected)
			}
		}
		t.Errorf(message)
		return false
	}

	expectedVal := reflect.ValueOf(expected)
	collectionVal := reflect.ValueOf(collection)

	// Case 1: Expected is a slice and collection is a single value (non-string)
	if (expectedVal.Kind() == reflect.Slice || expectedVal.Kind() == reflect.Array) &&
		collectionVal.Kind() != reflect.Slice && collectionVal.Kind() != reflect.Array {

		// If collection is a string, check if any element from expected is contained in the string
		if collectionVal.Kind() == reflect.String {
			collectionStr := collection.(string)
			allStrings := true

			for i := 0; i < expectedVal.Len(); i++ {
				if _, ok := expectedVal.Index(i).Interface().(string); !ok {
					allStrings = false
					break
				}
			}

			if allStrings {
				for i := 0; i < expectedVal.Len(); i++ {
					elemStr := expectedVal.Index(i).Interface().(string)
					if strings.Contains(collectionStr, elemStr) {
						return true
					}
				}

				message := fmt.Sprintf("Expected string '%s' to contain at least one element from %v",
					collectionStr, expected)
				if len(msgAndArgs) > 0 {
					if msgFormat, ok := msgAndArgs[0].(string); ok {
						customMsg := formatMessage(msgFormat, msgAndArgs[1:]...)
						message = fmt.Sprintf("%s (%s)", customMsg, message)
					}
				}
				t.Errorf(message)
				return false
			}
		}

		// Compare each element in the slice to the collection value
		for i := 0; i < expectedVal.Len(); i++ {
			if reflect.DeepEqual(expectedVal.Index(i).Interface(), collection) {
				return true
			}
		}

		message := fmt.Sprintf("Expected value %v to equal at least one element from %v",
			collection, expected)
		if len(msgAndArgs) > 0 {
			if msgFormat, ok := msgAndArgs[0].(string); ok {
				customMsg := formatMessage(msgFormat, msgAndArgs[1:]...)
				message = fmt.Sprintf("%s (%s)", customMsg, message)
			}
		}
		t.Errorf(message)
		return false
	}

	// Case 2: Expected is a string and collection is a slice of strings
	if expectedStr, ok := expected.(string); ok && (collectionVal.Kind() == reflect.Slice || collectionVal.Kind() == reflect.Array) {
		allStrings := true
		for i := 0; i < collectionVal.Len(); i++ {
			if _, ok := collectionVal.Index(i).Interface().(string); !ok {
				allStrings = false
				break
			}
		}

		if allStrings {
			missingChars := []rune{}
			for _, char := range expectedStr {
				found := false
				for i := 0; i < collectionVal.Len(); i++ {
					elemStr := collectionVal.Index(i).Interface().(string)
					if strings.Contains(elemStr, string(char)) {
						found = true
						break
					}
				}
				if !found {
					missingChars = append(missingChars, char)
				}
			}

			if len(missingChars) == 0 {
				return true
			}

			message := fmt.Sprintf("Expected slice to contain all characters from '%s', missing: %v",
				expectedStr, string(missingChars))
			if len(msgAndArgs) > 0 {
				if msgFormat, ok := msgAndArgs[0].(string); ok {
					customMsg := formatMessage(msgFormat, msgAndArgs[1:]...)
					message = fmt.Sprintf("%s (%s)", customMsg, message)
				}
			}
			t.Errorf(message)
			return false
		}
	}

	// Fall back to Contains for other cases
	return Contains(t, expected, collection, msgAndArgs...)
}

// NotEmpty verifies that the collection (slice, map, string, etc.) is not empty.
// If it is empty, it formats an error message and reports it through the test interface.
func NotEmpty(t interface {
	Errorf(format string, args ...any)
	Helper()
}, collection any, msgAndArgs ...any) bool {
	t.Helper()
	
	if collection == nil {
		message := formatEmptyMessage("nil value", msgAndArgs...)
		t.Errorf(message)
		return false
	}

	v := reflect.ValueOf(collection)

	// Handle typed nil slices, maps, etc.
	if (v.Kind() == reflect.Slice || v.Kind() == reflect.Map ||
		v.Kind() == reflect.Chan) && v.IsNil() {
		message := formatEmptyMessage(fmt.Sprintf("nil %v", v.Type()), msgAndArgs...)
		t.Errorf(message)
		return false
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String, reflect.Chan:
		if v.Len() > 0 {
			return true
		}
		message := formatEmptyMessage(fmt.Sprintf("empty %v", v.Type()), msgAndArgs...)
		t.Errorf(message)
		return false
	default:
		// For non-collection types, check if it's the zero value
		isZero := reflect.ValueOf(collection).IsZero()
		if !isZero {
			return true
		}
	}

	message := formatEmptyMessage(fmt.Sprintf("zero value of type %T", collection), msgAndArgs...)
	t.Errorf(message)
	return false
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

// formatEmptyMessage creates a descriptive message for empty values
func formatEmptyMessage(valueDesc string, msgAndArgs ...any) string {
	var msg string
	if len(msgAndArgs) > 0 {
		if msgFormat, ok := msgAndArgs[0].(string); ok {
			message := formatMessage(msgFormat, msgAndArgs[1:]...)
			msg = message
			return fmt.Sprintf("%s (Got: %s)", msg, valueDesc)
		}
	}
	return fmt.Sprintf("Expected non-empty value, got: %s", valueDesc)
}
