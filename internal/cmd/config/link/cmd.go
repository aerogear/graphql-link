package new

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/chirino/graphql-gw/internal/cmd/config"
	"github.com/chirino/graphql-gw/internal/cmd/root"
	"github.com/chirino/graphql-gw/internal/gateway"
	"github.com/chirino/graphql/schema"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	Command = &cobra.Command{
		Use:   "link [upstream] [type] [field]",
		Short: "link an upstream into the gateway schema",
		Args:  cobra.ExactArgs(3),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			upstream = args[0]
			schemaType = args[1]
			field = args[2]
			return config.PreRunLoad(cmd, args)
		},
		Run: run,
	}
	upstream    string
	query       string
	schemaType  string
	field       string
	description string
	vars        []string
)

func init() {
	Command.Flags().StringVar(&description, "description", "", "description to add to the field (shown when introspected)")
	Command.Flags().StringSliceVar(&vars, "var", []string{}, "variable fields to extract from the current type in '$[name]=[query]' format")
	Command.Flags().StringVar(&query, "query", "query {}", "a partial graphql query what the root path to mount from the upstream server")
	config.Command.AddCommand(Command)
}

func run(_ *cobra.Command, _ []string) {

	c := config.Value
	log := c.Log

	if _, ok := c.Config.Upstreams[upstream]; !ok {
		log.Fatalf("upstream %s not found in the configuration", upstream)
	}

	document := schema.QueryDocument{}
	err := document.ParseWithDescriptions(query)
	if err != nil {
		log.Fatalf("invalid query argument: "+root.Verbosity, err)
	}

	gw, err := gateway.New(c.Config)
	if err != nil {
		log.Fatalf("existing gateway configuration is invalid: "+root.Verbosity, err)
	}

	if gw.Schema.Types[schemaType] == nil {
		log.Fatalf("gateway does not curretly have type named: %s", schemaType)
	}

	varMap := map[string]string{}
	for _, s := range vars {
		parts := strings.SplitN(s, "=", 2)
		if len(parts) != 2 {
			log.Fatalf("invalid --var syntax '%s'", s)
		}

		document := schema.QueryDocument{}
		err := document.ParseWithDescriptions("{" + parts[1] + "}")
		if err != nil {
			log.Fatalf("invalid --var query "+root.Verbosity, err)
		}
		varMap[parts[0]] = parts[1]
	}

	byName := map[string]*gateway.TypeConfig{}
	for _, t := range c.Types {
		existing := byName[t.Name]
		if existing != nil {
			existing.Actions = append(existing.Actions, t.Actions...)
		} else {
			byName[t.Name] = &t
		}
	}

	existing := byName[schemaType]
	if existing == nil {
		existing = &gateway.TypeConfig{Name: schemaType}
		byName[schemaType] = existing
	}

	existing.Actions = append(existing.Actions, gateway.ActionWrapper{
		Action: &gateway.Link{
			Field:       field,
			Description: description,
			Upstream:    upstream,
			Query:       query,
			Vars:        varMap,
		},
	})

	c.Types = []gateway.TypeConfig{}
	for _, t := range byName {
		c.Types = append(c.Types, *t)
	}

	configYml, err := yaml.Marshal(c)
	configFile := filepath.Join("./", config.File)
	err = ioutil.WriteFile(configFile, configYml, 0644)
	if err != nil {
		log.Fatalf(root.Verbosity, err)
	}
	log.Printf(`link added`)
}
