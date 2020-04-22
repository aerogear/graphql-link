package serve

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"

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
)

func init() {
	Command.Flags().StringVar(&ConfigFile, "config", "graphql-gw.yaml", "path to the config file to load")
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
		log.Fatalf(vebosityFmt, err)
	}

	config := Config{}
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		log.Fatalf(vebosityFmt, err)
	}

	if config.Listen == "" {
		config.Listen = "0.0.0.0:8080"
	}

	engine, err := gateway.NewEngine(config.Config)
	if err != nil {
		log.Fatalf(vebosityFmt, err)
	}

	host, port, err := net.SplitHostPort(config.Listen)
	if err != nil {
		log.Fatalf(vebosityFmt, err)
	}

	http.Handle("/graphql", &relay.Handler{Engine: engine})
	log.Printf("GraphQL endpoint running at http://%s:%s/graphql\n", host, port)

	http.Handle("/", graphiql.New("http://localhost:8080/graphql", false))
	log.Printf("GraphQL UI running at http://%s:%s/\n", host, port)
	log.Fatalf(vebosityFmt, http.ListenAndServe(config.Listen, nil))
}
