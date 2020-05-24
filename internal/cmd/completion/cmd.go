package completion

import (
	"os"

	"github.com/chirino/graphql-gw/internal/cmd/root"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "completion",
		Short: "Generates bash completion scripts",
		Long: `To load completion run

. <(graphql-gw completion)

To configure your bash shell to load completions for each session run:

echo ". <(graphql-gw completion)" >> ~/.profile

`,
		Run: func(cmd *cobra.Command, args []string) {
			root.Command.GenBashCompletion(os.Stdout)
		},
	}
)

func init() {
	root.Command.AddCommand(Command)
}
