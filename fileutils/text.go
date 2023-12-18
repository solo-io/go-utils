package fileutils

import "os"

func ReadFileString(filename string) (string, error) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}
