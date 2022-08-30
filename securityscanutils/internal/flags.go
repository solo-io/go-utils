package internal

import "github.com/spf13/pflag"

type GlobalFlags struct {
	Verbose bool
}

func (g *GlobalFlags) AddToFlags(flags *pflag.FlagSet) {
	flags.BoolVarP(&g.Verbose, "verbose", "v", false, "Enable verbose logging")
}
