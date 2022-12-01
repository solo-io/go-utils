package cliutils

import (
	errors "github.com/rotisserie/eris"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type HideableFlag interface {
	MarkHidden(string) error
}

// MustMarkHidden panics if the call to MarkHidden() fails.
func MustMarkHidden(flags HideableFlag, name string) {
	if err := flags.MarkHidden(name); err != nil {
		panic(err)
	}
}

// MustMarkFlagRequired panics if the call to MarkFlagRequired() fails.
func MustMarkFlagRequired(flaggish interface{}, name string) {
	if err := MarkFlagRequired(flaggish, name); err != nil {
		panic(err)
	}
}

// MarkFlagRequired returns an error if the call to MarkFlagRequired() fails.
func MarkFlagRequired(flaggish interface{}, name string) error {
	switch v := flaggish.(type) {
	case *cobra.Command:
		if err := v.MarkFlagRequired(name); err != nil {
			return err
		}
	case *pflag.FlagSet:
		if err := cobra.MarkFlagRequired(v, name); err != nil {
			return err
		}

	default:
		return errors.Errorf("Unknown flag type: %v", v)
	}
	return nil
}

// MustMarkPersistentFlagRequired panics if the call to MarkPersistentFlagRequired() fails.
func MustMarkPersistentFlagRequired(cmd *cobra.Command, name string) {
	if err := cmd.MarkPersistentFlagRequired(name); err != nil {
		panic(err)
	}
}
