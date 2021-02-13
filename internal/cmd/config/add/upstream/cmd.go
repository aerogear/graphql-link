package new

import (
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/aerogear/graphql-link/internal/cmd/config"
	"github.com/aerogear/graphql-link/internal/cmd/config/add"
	"github.com/aerogear/graphql-link/internal/cmd/root"
	"github.com/aerogear/graphql-link/internal/gateway"
	"github.com/chirino/graphql"
	"github.com/chirino/graphql/schema"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "upstream [name] [url]",
		Short: "Adds new upstream to config.",
		Long:  `Command lets you assemble gateway config by letting you add new upstream gateway`,
		Args:  cobra.ExactArgs(2),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			upstreamName = args[0]
			upstreamURL = args[1]
			return config.PreRunLoad(cmd, args)
		},
		Run: run,
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
	upstreamServer, err := gateway.CreateGraphQLUpstreamServer(upstreamName, upstream)
	if err != nil {
		log.Fatalf(root.Verbosity, "invalid upstream: "+root.Verbosity, err)
	}

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
	os.MkdirAll(filepath.Join(c.WorkDirectory, "upstreams"), 0755)
	upstreamSchemaFile := filepath.Join(c.WorkDirectory, "upstreams", upstreamName+".graphql")
	err = ioutil.WriteFile(upstreamSchemaFile, []byte(upstreamSchema.String()), 0644)
	if err != nil {
		log.Fatalf(root.Verbosity, err)
	}

	// Update the config
	c.Upstreams[upstreamName] = gateway.UpstreamWrapper{Upstream: upstream}

	err = config.Store(*c)
	if err != nil {
		log.Fatalf(root.Verbosity, err)
	}
	log.Printf(`upstream added`)
}
