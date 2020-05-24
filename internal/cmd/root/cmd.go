package root

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "graphql-gw",
		Short: "A GraphQL composition gateway",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if !Verbose {
				Verbosity = "%+v"
			}
		},
	}
	Verbose   = false
	Verbosity = "%v"
)

func init() {
	Command.PersistentFlags().BoolVar(&Verbose, "verbose", false, "enables increased verbosity")
}

func Main() {
	if err := Command.Execute(); err != nil {
		fmt.Printf(Verbosity+"\n", err)
		os.Exit(1)
	}
}
