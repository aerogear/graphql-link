package root

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	Command = &cobra.Command{
		Use:   "graphql-link",
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
	viper.SetEnvPrefix("GRAPHQL_GW")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

func Main() {
	if err := Command.Execute(); err != nil {
		fmt.Printf(Verbosity+"\n", err)
		os.Exit(1)
	}
}

func IntBind(f *pflag.FlagSet, name string, defaultValue int, usage string) func() int {
	f.Int(name, defaultValue, usage)
	viper.SetDefault(name, defaultValue)
	viper.BindPFlag(name, f.Lookup(name))
	return func() int {
		return viper.GetInt(name)
	}
}

func StringBind(f *pflag.FlagSet, name string, defaultValue string, usage string) func() string {
	f.String(name, defaultValue, usage)
	viper.SetDefault(name, defaultValue)
	viper.BindPFlag(name, f.Lookup(name))
	return func() string {
		return viper.GetString(name)
	}
}

func BoolBind(f *pflag.FlagSet, name string, defaultValue bool, usage string) func() bool {
	f.Bool(name, defaultValue, usage)
	viper.SetDefault(name, defaultValue)
	viper.BindPFlag(name, f.Lookup(name))
	return func() bool {
		return viper.GetBool(name)
	}
}
