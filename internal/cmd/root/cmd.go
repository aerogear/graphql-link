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
	}
	Verbose = false
)

func init() {
	Command.PersistentFlags().BoolVar(&Verbose, "verbose", false, "enables increased verbosity")
}

func Main() {
	if err := Command.Execute(); err != nil {
		if Verbose {
			fmt.Printf("%+v\n", err)
		} else {
			fmt.Printf("%v\n", err)
		}
		os.Exit(1)
	}
}
