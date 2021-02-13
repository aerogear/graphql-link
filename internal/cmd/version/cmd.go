package version

import (
	"fmt"

	"github.com/aerogear/graphql-link/internal/cmd/root"
	"github.com/spf13/cobra"
)

type VersionConfig struct {
	Version string
	Commit  string
	Date    string
}

var Config = VersionConfig{
	Version: "dev",
	Commit:  "none",
	Date:    "unknown",
}

var (
	Command = &cobra.Command{
		Use:   "version",
		Short: "Print version information for this executable",
		Run:   run,
	}
)

func init() {
	root.Command.AddCommand(Command)
}

func run(cmd *cobra.Command, args []string) {
	fmt.Println("version:", Config.Version)
	fmt.Println("commit:", Config.Commit)
	fmt.Println("date:", Config.Date)
}
