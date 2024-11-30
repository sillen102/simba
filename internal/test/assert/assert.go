package assert

import (
	"fmt"
	"reflect"
	"testing"
)

// Equal asserts that two values are equal
func Equal(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("\nexpected: %v\nactual  : %v", expected, actual)
	}
}

// NoError asserts that the error is nil
func NoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

// Error asserts that an error occurred
func Error(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Error("expected an error, got nil")
	}
}

// ErrorIs asserts that an error matches the expected error
func ErrorIs(t *testing.T, err, target error) {
	t.Helper()
	if !ErrorsIs(err, target) {
		t.Errorf("expected error %v, got %v", target, err)
	}
}

// ErrorsIs is a helper function that reports whether err matches target
func ErrorsIs(err, target error) bool {
	return fmt.Sprint(err) == fmt.Sprint(target)
}
