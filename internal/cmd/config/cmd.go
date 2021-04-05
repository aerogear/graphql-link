package config

import (
	"io/ioutil"
	"net/http/httptest"
	"path/filepath"

	"github.com/aerogear/graphql-link/internal/cmd/root"
	"github.com/aerogear/graphql-link/internal/gateway"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Server         *httptest.Server `yaml:"-"`
	Gateway        *gateway.Gateway `yaml:"-"`
	Listen         string           `yaml:"listen"`
	gateway.Config `yaml:"-,inline"`
}

var (
	Command = &cobra.Command{
		Use:               "config",
		Short:             "Modifies the gateway configuration",
		PersistentPreRunE: PreRunLoad,
	}
	File    string
	WorkDir string
	Value   *Config
)

func init() {
	Command.PersistentFlags().StringVar(&File, "config", "graphql-link.yaml", "path to the config file to modify")
	Command.PersistentFlags().StringVar(&WorkDir, "workdir", "", "working to write files to in dev mode. (default to the directory the config file is in)")
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

	if WorkDir == "" {
		config.WorkDirectory = filepath.Dir(File)
	} else {
		config.WorkDirectory = WorkDir
	}

	config.Log = gateway.SimpleLog

	if config.Upstreams == nil {
		config.Upstreams = map[string]gateway.UpstreamWrapper{}
	}
	if config.Types == nil {
		config.Types = []gateway.TypeConfig{}
	}

	return nil
}

func Store(config Config) error {
	configYml, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(File, configYml, 0644)
	if err != nil {
		return err
	}
	return nil
}
