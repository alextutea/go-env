package errs

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

type RequiredKeyNotPresentError struct {
	Keys []string
}

func (e RequiredKeyNotPresentError) Error() string {
	if len(e.Keys) == 1 {
		return fmt.Sprintf(
			"required env key missing; %s must be provided", e.Keys[0],
		)
	}
	var sb strings.Builder
	for i, key := range e.Keys {
		sb.WriteString(key)
		if i != len(e.Keys)-1 {
			sb.WriteString(", ")
		}
	}
	return fmt.Sprintf(
		"required env key missing; one of the following keys must be provided [%s]",
		sb.String(),
	)
}

func NewRequiredKeyNotPresentError(keys ...string) RequiredKeyNotPresentError {
	return RequiredKeyNotPresentError{
		Keys: keys,
	}
}

func IsRequiredKeyNotPresentError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := errors.Cause(err).(RequiredKeyNotPresentError)
	return ok
}
