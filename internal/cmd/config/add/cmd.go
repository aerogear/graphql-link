package add

import (
	"github.com/chirino/graphql-gw/internal/cmd/config"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "add",
		Short: "adds an upstream server",
	}
)

func init() {
	config.Command.AddCommand(Command)
}
