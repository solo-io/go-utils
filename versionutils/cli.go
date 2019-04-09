package versionutils

import (
	"fmt"
	"os"
)

func GetReleaseVersionOrExitGracefully() *Version {
	tag, present := os.LookupEnv("TAGGED_VERSION")
	if !present || tag == "" {
		fmt.Printf("TAGGED_VERSION not found in environment.\n")
		os.Exit(0)
	}
	version, err := ParseVersion(tag)
	if err != nil {
		fmt.Printf("TAGGED_VERSION %s is not a valid semver version.\n", tag)
		os.Exit(0)
	}
	return version
}
