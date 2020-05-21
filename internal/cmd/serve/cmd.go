package serve

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/chirino/graphql-gw/internal/cmd/root"
	"github.com/chirino/graphql-gw/internal/gateway"
	"github.com/chirino/graphql/graphiql"
	"github.com/chirino/graphql/httpgql"
	"github.com/fsnotify/fsnotify"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
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
	Listen    string `json:"listen"`
	verbosity string `json:"-"`
}

func run(cmd *cobra.Command, args []string) {
	config := Config{}
	config.ConfigDirectory = filepath.Dir(ConfigFile)
	config.Log = gateway.SimpleLog
	config.verbosity = "%v"
	if !root.Verbose {
		config.verbosity = "%+v\n"
	}

	fileConfig := config
	err := readConfig(&fileConfig)
	if err != nil {
		config.Log.Fatalf("error reading configuration file: "+config.verbosity, err)
	}

	stopper := startServer(fileConfig)
	restart := func() {

		fileConfig := config
		err := readConfig(&fileConfig)
		if err != nil {
			config.Log.Fatalf("error reading configuration file: "+config.verbosity, err)
		}

		stopper()
		if err != nil {
			config.Log.Fatalf("error occured while stopping server: "+config.verbosity, err)
		}

		stopper = startServer(fileConfig)
	}

	if !Production {
		watchFile(ConfigFile, func(in fsnotify.Event) {
			config.Log.Println("restarting due to configuration change:", in.Name)
			restart()
		})
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	for {
		switch <-sigs {
		case syscall.SIGINT:
			config.Log.Println("shutting down due to SIGINT signal")
			stopper()
			os.Exit(0)

		case syscall.SIGTERM:
			config.Log.Println("shutting down due to SIGTERM signal")
			stopper()
			os.Exit(0)

		case syscall.SIGHUP:
			config.Log.Println("restarting due to SIGHUP signal")
			restart()
		}
	}
}

func watchFile(filename string, onChange func(in fsnotify.Event)) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	realConfigFile, _ := filepath.EvalSymlinks(filename)
	err = watcher.Add(realConfigFile)
	if err != nil {
		watcher.Close()
		return nil, err
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				const writeOrCreateMask = fsnotify.Write | fsnotify.Create
				if event.Op&writeOrCreateMask != 0 {
					onChange(event)
				}
			case <-watcher.Errors:
				return
			}
		}
	}()
	return watcher, nil
}

func readConfig(config interface{}) error {
	file, err := ioutil.ReadFile(ConfigFile)

	if err != nil {
		return errors.Wrapf(err, "reading config file: %s.", ConfigFile)
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return errors.Wrapf(err, "parsing yaml of: %s.", ConfigFile)
	}
	return nil
}

func startServer(config Config) (stopper func()) {

	// Let's only apply the env expansion to the URLs for now.
	// We don't want to run it on queries which can have $var expressions
	// in them.
	for _, ep := range config.Upstreams {
		switch upstream := ep.Upstream.(type) {
		case *gateway.GraphQLUpstream:
			upstream.URL = os.ExpandEnv(upstream.URL)
		}
	}

	if config.Listen == "" {
		config.Listen = "0.0.0.0:8080"
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
		config.Log.Fatalf(config.verbosity, err)
	}

	mux := http.NewServeMux()
	server, err := gateway.StartHttpListener(config.Listen, mux)
	if err != nil {
		config.Log.Fatalf(config.verbosity, err)
	}

	gatewayHandler := gateway.CreateHttpHandler(engine.ServeGraphQLStream).(*httpgql.Handler)
	// Enable pretty printed json results when in dev mode.
	if !Production {
		gatewayHandler.Indent = "  "
	}

	graphqlURL := fmt.Sprintf("%s/graphql", server.URL)
	mux.Handle("/graphql", gatewayHandler)
	mux.Handle("/", graphiql.New(graphqlURL, true))

	config.Log.Printf("GraphQL endpoint running at %s", graphqlURL)
	config.Log.Printf("GraphQL UI running at %s", server.URL)

	return server.Close
}
