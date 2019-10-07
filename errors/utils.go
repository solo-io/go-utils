package errors

import (
	"errors"
	"fmt"
	"strings"
)

func Wrapf(err error, format string, args ...interface{}) error {
	errString := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", errString, err)
}

func Errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

func Errors(msgs []string) error {
	return fmt.Errorf(strings.Join(msgs, "\n"))
}

func New(text string) error {
	return errors.New(text)
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}
