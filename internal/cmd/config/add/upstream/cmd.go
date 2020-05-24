package new

import (
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql-gw/internal/cmd/config"
	"github.com/chirino/graphql-gw/internal/cmd/config/add"
	"github.com/chirino/graphql-gw/internal/cmd/root"
	"github.com/chirino/graphql-gw/internal/gateway"
	"github.com/chirino/graphql/schema"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "upstream [name] [url]",
		Short: "Adds new upstream to config.",
		Long:  `Command lets you assemble gateway config by letting you add new upstream gateway`,
		Run:   run,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			upstreamName = args[0]
			upstreamURL = args[1]
		},
	}
	upstreamName = ""
	upstreamURL  = ""
	prefix       = ""
	suffix       = ""
	schemaFile   = ""
)

func init() {
	Command.Flags().StringVar(&prefix, "prefix", "", "rename all upstream types with the prefix")
	Command.Flags().StringVar(&suffix, "suffix", "", "rename all upstream types with the suffix")
	Command.Flags().StringVar(&schemaFile, "schema-file", "", "path to schema file of the server (default downloads the schema from the upstream via introspection)")
	add.Command.AddCommand(Command)
}

func run(cmd *cobra.Command, args []string) {
	c := config.Value
	log := c.Log

	_, err := url.Parse(upstreamURL)
	if err != nil {
		log.Fatalf("error reading config file: "+root.Verbosity, err)
	}

	gw, err := gateway.New(c.Config)
	if err != nil {
		log.Fatalf(root.Verbosity, "existing gateway configuration is invalid: "+root.Verbosity, err)
	}

	if _, ok := c.Upstreams[upstreamName]; ok {
		log.Fatalf("an upstream named '%s' already exists", upstreamName)
	}

	upstream := &gateway.GraphQLUpstream{
		URL:    upstreamURL,
		Prefix: prefix,
		Suffix: suffix,
	}

	var upstreamSchema *schema.Schema
	upstreamServer := gateway.CreateUpstreamServer(upstreamName, upstream)

	if schemaFile == "" {
		// Verify we can get the schema from the server...
		upstreamSchema, err = graphql.GetSchema(upstreamServer.Client)
		if err != nil {
			log.Fatalf(root.Verbosity, "could not download the upstream schema: "+root.Verbosity, err)
		}

	} else {

		file, err := ioutil.ReadFile(schemaFile)
		if err != nil {
			log.Fatalf("error reading schema file: "+root.Verbosity, err)
		}

		upstreamSchema = &schema.Schema{}
		err = upstreamSchema.Parse(string(file))
		if err != nil {
			log.Fatalf("bad schema: "+root.Verbosity, err)
		}
	}
	upstreamServer.RenameTypes(upstreamSchema)

	// Verify none of the upstream types conflict with the exiting gateway types.
	for _, t := range upstreamServer.Schema.Types {
		name := t.TypeName()
		if schema.Meta.Types[name] != nil {
			continue
		}

		if gw.Schema.Types[name] != nil {
			log.Fatalf("upstream type '%s' already exists in the gatway schema, try use the --prefix or --suffix options to automatically rename the upstream types", name)
		}
	}

	// Store it's schema
	os.MkdirAll(filepath.Join(c.ConfigDirectory, "upstreams"), 0755)
	upstreamSchemaFile := filepath.Join(c.ConfigDirectory, "upstreams", upstreamName+".graphql")
	err = ioutil.WriteFile(upstreamSchemaFile, []byte(upstreamSchema.String()), 0644)
	if err != nil {
		log.Fatalf(root.Verbosity, err)
	}

	// Update the config

	if c.Upstreams == nil {
		c.Upstreams = map[string]gateway.UpstreamWrapper{}
	}
	if c.Types == nil {
		c.Types = []gateway.TypeConfig{}
	}
	c.Upstreams[upstreamName] = gateway.UpstreamWrapper{Upstream: upstream}

	configYml, err := yaml.Marshal(&c)
	configFile := filepath.Join("./", config.File)
	err = ioutil.WriteFile(configFile, configYml, 0644)
	if err != nil {
		log.Fatalf(root.Verbosity, err)
	}

	log.Printf(`upstream added`)
}
