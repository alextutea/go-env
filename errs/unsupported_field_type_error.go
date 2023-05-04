package errs

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
)

type UnsupportedFieldTypeError struct {
	ActualType reflect.Type
}

func (e UnsupportedFieldTypeError) Error() string {
	return fmt.Sprintf("all fields must be a string, int, float, bool or struct, not a %s", e.ActualType.Name())
}

func NewUnsupportedFieldTypeError(t reflect.Type) UnsupportedFieldTypeError {
	return UnsupportedFieldTypeError{
		ActualType: t,
	}
}

func IsUnsupportedFieldTypeError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := errors.Cause(err).(UnsupportedFieldTypeError)
	return ok
}
