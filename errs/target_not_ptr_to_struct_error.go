package errs

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
)

type TargetNotPtrToStructError struct {
	ActualType reflect.Type
}

func (e TargetNotPtrToStructError) Error() string {
	return fmt.Sprintf("Target passed must be a reference to a struct type, not a %s", e.ActualType.Name())
}

func NewTargetNotPtrToStructError(t reflect.Type) TargetNotPtrToStructError {
	return TargetNotPtrToStructError{
		ActualType: t,
	}
}

func IsTargetNotPtrToStructError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := errors.Cause(err).(TargetNotPtrToStructError)
	return ok
}
