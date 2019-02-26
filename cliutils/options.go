package cliutils

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

type Options interface {
	Initialize()
}

type OptionsFunc = func(*cobra.Command)
type CmdFunc = func(Options, ...OptionsFunc) *cobra.Command

func ApplyOptions(cmd *cobra.Command, funcs []OptionsFunc) {
	for _, v := range funcs {
		v(cmd)
	}
}

func ReplaceCmd(parent *cobra.Command, new *cobra.Command) error {
	parentCmds := parent.Commands()
	for _, old := range parentCmds {
		if old.Use == new.Use {
			parent.RemoveCommand(old)
			parent.AddCommand(new)
			return nil
		}
	}
	return fmt.Errorf("did not find child command to replace")
}

func MustReplaceCmd(parent *cobra.Command, new *cobra.Command) {
	if err := ReplaceCmd(parent, new); err != nil {
		log.Fatal(err)
	}
}
