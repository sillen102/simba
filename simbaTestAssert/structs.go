package simbaTestAssert

import (
	"fmt"
	"reflect"
)

// Equal compares two values of the same type and returns true if they're equal.
// If they're not equal, it formats an error message and reports it through the test interface.
func Equal[T any](t interface {
	Errorf(format string, args ...interface{})
}, expected, actual T, msgAndArgs ...interface{}) bool {
	if reflect.DeepEqual(expected, actual) {
		return true
	}

	message := formatFailureMessage(expected, actual, msgAndArgs...)
	t.Errorf("%s", message)
	return false
}

// NotEqual compares two values of the same type and returns true if they're not equal.
// If they're equal, it formats an error message and reports it through the test interface.
func NotEqual[T any](t interface {
	Errorf(format string, args ...interface{})
}, expected, actual T, msgAndArgs ...interface{}) bool {
	if !reflect.DeepEqual(expected, actual) {
		return true
	}

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

	t.Errorf("%sExpected values to not be equal:\n%#v", msg, expected)
	return false
}

// NilError checks if the error is nil and returns true if it is.
// If the error is not nil, it formats an error message and reports it through the test interface.
func NilError(t interface {
	Errorf(format string, args ...interface{})
}, err error, msgAndArgs ...interface{}) bool {
	if err == nil {
		return true
	}

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

	t.Errorf("%sExpected nil error, got: %v", msg, err)
	return false
}

// formatFailureMessage creates a descriptive failure message for unequal values
func formatFailureMessage(expected, actual interface{}, msgAndArgs ...interface{}) string {
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

	return fmt.Sprintf("%sExpected: %#v\nActual:   %#v",
		msg,
		expected,
		actual)
}
