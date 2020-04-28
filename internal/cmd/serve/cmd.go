package serve

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/chirino/graphql-gw/internal/cmd/root"
	"github.com/chirino/graphql-gw/internal/gateway"
	"github.com/chirino/graphql/graphiql"
	"github.com/chirino/graphql/relay"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "serve",
		Short: "Runs the gateway service",
		Run:   run,
	}
	ConfigFile = ""
	Production = false
)

func init() {
	Command.Flags().StringVar(&ConfigFile, "config", "graphql-gw.yaml", "path to the config file to load")
	Command.Flags().BoolVar(&Production, "production", false, "when true, the server will not download and store schemas from remote graphql endpoints.")
	root.Command.AddCommand(Command)
}

type Config struct {
	gateway.Config
	Listen string `json:"listen"`
}

func run(cmd *cobra.Command, args []string) {
	vebosityFmt := "%v"
	if !root.Verbose {
		vebosityFmt = "%+v\n"
	}

	file, err := ioutil.ReadFile(ConfigFile)

	if err != nil {
		log.Fatalf("Error reading config file: %s.", err)
		panic(err)
	}

	file = []byte(os.ExpandEnv(string(file)))
	config := Config{}
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		log.Fatalf("Error parsing yaml file: %s.", err)
		panic(err)
	}

	if config.Listen == "" {
		config.Listen = "localhost:8080"
	}

	if Production {
		config.DisableSchemaDownloads = true
		config.EnabledSchemaStorage = false
	} else {
		config.DisableSchemaDownloads = false
		config.EnabledSchemaStorage = true
	}

	engine, err := gateway.New(config.Config)
	if err != nil {
		log.Fatalf(vebosityFmt, err)
	}

	host, port, err := net.SplitHostPort(config.Listen)
	if err != nil {
		log.Fatalf(vebosityFmt, err)
	}

	endpoint := fmt.Sprintf("http://%s:%s", host, port)
	http.Handle("/graphql", &relay.Handler{ServeGraphQLStream: engine.ServeGraphQLStream})
	log.Printf("GraphQL endpoint running at %s/graphql", endpoint)
	http.Handle("/", graphiql.New(endpoint+"/graphql", false))
	log.Printf("GraphQL UI running at %s", endpoint)

	log.Fatalf(vebosityFmt, http.ListenAndServe(config.Listen, nil))
}
