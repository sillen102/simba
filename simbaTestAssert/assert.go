package simbaTestAssert

import (
	"fmt"
	"reflect"
)

// True checks if the condition is true and returns true if it is.
func True(t interface {
	Errorf(format string, args ...any)
}, condition bool, msgAndArgs ...any) bool {
	if condition {
		return true
	}

	message := "Expected true, got false"
	if len(msgAndArgs) > 0 {
		if msgFormat, ok := msgAndArgs[0].(string); ok {
			message = formatMessage(msgFormat, msgAndArgs[1:]...)
		}
	}

	t.Errorf(message)
	return false
}

// False checks if the condition is false and returns true if it is.
func False(t interface {
	Errorf(format string, args ...any)
}, condition bool, msgAndArgs ...any) bool {
	if !condition {
		return true
	}

	message := "Expected false, got true"
	if len(msgAndArgs) > 0 {
		if msgFormat, ok := msgAndArgs[0].(string); ok {
			message = formatMessage(msgFormat, msgAndArgs[1:]...)
		}
	}

	t.Errorf(message)
	return false
}

// Equal compares two values of the same type and returns true if they're equal.
// If they're not equal, it formats an error message and reports it through the test interface.
func Equal[T any](t interface {
	Errorf(format string, args ...any)
}, expected, actual T, msgAndArgs ...any) bool {
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
	Errorf(format string, args ...any)
}, expected, actual T, msgAndArgs ...any) bool {
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

// NoError checks if the error is nil and returns true if it is.
// If the error is not nil, it formats an error message and reports it through the test interface.
func NoError(t interface {
	Errorf(format string, args ...any)
}, err error, msgAndArgs ...any) bool {
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

// Error checks if the error is not nil and returns true if it is not.
func Error(t interface {
	Errorf(format string, args ...any)
}, err error, msgAndArgs ...any) bool {
	if err != nil {
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

	t.Errorf("%sExpected non-nil error, got: nil", msg)
	return false
}

// Assert verifies that the condition is true.
// If not, it formats an error message and reports it through the test interface.
func Assert(t interface {
	Errorf(format string, args ...any)
}, condition bool, msgAndArgs ...any) bool {
	if condition {
		return true
	}

	message := "Assertion failed"
	if len(msgAndArgs) > 0 {
		if msgFormat, ok := msgAndArgs[0].(string); ok {
			message = formatMessage(msgFormat, msgAndArgs[1:]...)
		} else {
			message = fmt.Sprintf("Assertion failed: %+v", msgAndArgs[0])
		}
	}

	t.Errorf(message)
	return false
}

// Nil checks if the value is nil and returns true if it is.
// If the value is not nil, it formats an error message and reports it through the test interface.
func Nil(t interface {
	Errorf(format string, args ...any)
}, value any, msgAndArgs ...any) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	kind := v.Kind()

	// Handle typed nil interfaces and slices
	if (kind == reflect.Interface || kind == reflect.Slice || kind == reflect.Map ||
		kind == reflect.Chan || kind == reflect.Func || kind == reflect.Ptr) && v.IsNil() {
		return true
	}

	message := "Expected nil, got non-nil value"
	if len(msgAndArgs) > 0 {
		if msgFormat, ok := msgAndArgs[0].(string); ok {
			message = formatMessage(msgFormat, msgAndArgs[1:]...)
		}
	}

	finalMessage := fmt.Sprintf("%s: %#v", message, value)
	t.Errorf(finalMessage)
	return false
}

// NotNil checks if the value is not nil and returns true if it is not.
// If the value is nil, it formats an error message and reports it through the test interface.
func NotNil(t interface {
	Errorf(format string, args ...any)
}, value any, msgAndArgs ...any) bool {
	if value != nil && !(reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil()) {
		return true
	}

	message := "Expected non-nil value, got nil"
	if len(msgAndArgs) > 0 {
		if msgFormat, ok := msgAndArgs[0].(string); ok {
			message = formatMessage(msgFormat, msgAndArgs[1:]...)
		}
	}

	t.Errorf(message)
	return false
}

// formatFailureMessage creates a descriptive failure message for unequal values
func formatFailureMessage(expected, actual any, msgAndArgs ...any) string {
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

// formatMessage creates a formatted message from a format string and arguments
func formatMessage(format string, args ...any) string {
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}
