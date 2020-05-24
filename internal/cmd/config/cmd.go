package config

import (
	"io/ioutil"
	"net/http/httptest"
	"path/filepath"

	"github.com/chirino/graphql-gw/internal/cmd/root"
	"github.com/chirino/graphql-gw/internal/gateway"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type Config struct {
	gateway.Config
	Listen string `json:"listen"`
	Server *httptest.Server
}

var (
	Command = &cobra.Command{
		Use:               "config",
		Short:             "Modifies the gateway configuration",
		PersistentPreRunE: PreRunLoad,
	}
	File  string
	Value *Config
)

func init() {
	Command.PersistentFlags().StringVar(&File, "config", "graphql-gw.yaml", "path to the config file to modify")
	root.Command.AddCommand(Command)
}

func PreRunLoad(cmd *cobra.Command, args []string) error {
	Value = &Config{}
	return Load(Value)
}

func Load(config *Config) error {
	file, err := ioutil.ReadFile(File)

	if err != nil {
		return errors.Wrapf(err, "reading config file: %s.", File)
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return errors.Wrapf(err, "parsing yaml of: %s.", File)
	}

	config.ConfigDirectory = filepath.Dir(File)
	config.Log = gateway.SimpleLog
	return nil
}
