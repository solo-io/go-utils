package cliutils

import "github.com/spf13/cobra"

type Options interface {
	Initialize() error
}

type OptionsFunc func(*cobra.Command) error
type CmdFunc func(*Options, ...OptionsFunc) *cobra.Command