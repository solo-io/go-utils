package cliutils

import "github.com/spf13/cobra"

type OptionsFunc func(*cobra.Command)

func ApplyOptions(cmd *cobra.Command, funcs []OptionsFunc) {
	for _, v := range funcs {
		v(cmd)
	}
}
