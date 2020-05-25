package config

import (
	"io/ioutil"
	"net/http/httptest"
	"path/filepath"

	"github.com/chirino/graphql-gw/internal/cmd/root"
	"github.com/chirino/graphql-gw/internal/gateway"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Server         *httptest.Server `yaml:"-"`
	Listen         string           `yaml:"listen"`
	gateway.Config `yaml:"-,inline"`
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

	if config.Upstreams == nil {
		config.Upstreams = map[string]gateway.UpstreamWrapper{}
	}
	if config.Types == nil {
		config.Types = []gateway.TypeConfig{}
	}

	return nil
}
