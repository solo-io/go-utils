package osutils

import (
	"fmt"
	"os"

	"github.com/rotisserie/eris"
)

var (
	MissingEnvVarError = func(envVar string) error {
		return fmt.Errorf("Missing %s environment variable", envVar)
	}
)

/**
  Returns error if the environment variable is empty
*/
func GetEnvE(envVar string) (string, error) {
	env := os.Getenv(envVar)
	if env == "" {
		return "", MissingEnvVarError(envVar)
	}
	return env, nil
}

/**
  Creates dir if it doesn't already exist
  If the dir does exist, does nothing.
*/
func CreateDirIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return eris.Wrapf(err, "error creating dir %s", path)
		}
	}
	return nil
}
