package new

import (
	"github.com/chirino/graphql-gw/internal/cmd/config"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "mount [upstream] [query] (field)\n mount [upstream] [query] (field)",
		Short: "mount an upstream query path on the gateway schema",
		Run:   run,
	}
)

func init() {
	config.Command.AddCommand(Command)
}

func run(cmd *cobra.Command, args []string) {

	//config := root.CommandConfig
	//log := config.Log

}
