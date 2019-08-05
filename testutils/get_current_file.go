package testutils

import (
	"path/filepath"
	"runtime"

	"github.com/solo-io/go-utils/errors"
)

// returns the absolute path to the file the caller
// intended to provide a way to find test files
func GetCurrentFile() (string, error) {
	_, callerFile, _, ok := runtime.Caller(1)
	if !ok {
		return "", errors.Errorf("failed to get runtime.Caller")
	}
	return filepath.Abs(callerFile)
}
