package osutils

import (
	"fmt"
	"os"
)

var (
	MissingEnvVarError = func(envVar string) error {
		return fmt.Errorf("Missing %s environment variable", envVar)
	}
)

/*
*

	Returns error if the environment variable is empty
*/
func GetEnvE(envVar string) (string, error) {
	env := os.Getenv(envVar)
	if env == "" {
		return "", MissingEnvVarError(envVar)
	}
	return env, nil
}
