package serve

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

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
	verbosity string
	server    *httptest.Server
}

func run(cmd *cobra.Command, args []string) {
	config := Config{}
	config.ConfigDirectory = filepath.Dir(ConfigFile)
	config.Log = gateway.SimpleLog
	config.verbosity = "%v"
	if !root.Verbose {
		config.verbosity = "%+v\n"
	}

	lastConfig := config
	err := readConfig(&lastConfig)
	if err != nil {
		config.Log.Fatalf("error reading configuration file: "+config.verbosity, err)
	}

	err = lastConfig.startServer()
	if err != nil {
		config.Log.Fatalf("could not start the sever: "+config.verbosity, err)
	}

	restartMu := sync.Mutex{}
	restart := func() {
		restartMu.Lock()
		defer restartMu.Unlock()

		nextConfig := lastConfig
		err := readConfig(&nextConfig)
		if err != nil {
			config.Log.Fatalf("error reading configuration file: "+config.verbosity, err)
		}

		// restart the server on a port change...
		if lastConfig.Listen != nextConfig.Listen {
			lastConfig.server.Close()
			if err != nil {
				config.Log.Fatalf("error occured while stopping server: "+config.verbosity, err)
			}

			err = nextConfig.startServer()
			if err != nil {
				config.Log.Fatalf("could not start the sever: "+config.verbosity, err)
			}
		} else {
			// Just remount a new engine on to the server...
			nextConfig.postProcess()
			nextConfig.mountGatewayOnHttpServer()
		}

		lastConfig = nextConfig
	}

	if !Production {
		watchFile(ConfigFile, func(in fsnotify.Event) {
			config.Log.Println("restarting due to configuration change:", in.Name)
			restart()
		})

		go func() {
			for {
				changed, err := gateway.HaveUpstreamSchemaChanged(lastConfig.Config)
				if err == nil && changed {
					config.Log.Println("restarting due to an upstream schema change")
					restart()
				}
				time.Sleep(5 * time.Minute)
			}
		}()
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	for {
		switch <-sigs {
		case syscall.SIGINT:
			config.Log.Println("shutting down due to SIGINT signal")
			lastConfig.server.Close()
			os.Exit(0)

		case syscall.SIGTERM:
			config.Log.Println("shutting down due to SIGTERM signal")
			lastConfig.server.Close()
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

func (config *Config) startServer() error {
	config.postProcess()
	server, err := gateway.StartHttpListener(config.Listen, http.NewServeMux())
	if err != nil {
		return err
	}
	config.server = server
	config.mountGatewayOnHttpServer()
	return nil
}

func (config *Config) mountGatewayOnHttpServer() error {
	engine, err := gateway.New(config.Config)
	if err != nil {
		return err
	}
	gatewayHandler := gateway.CreateHttpHandler(engine.ServeGraphQLStream).(*httpgql.Handler)
	// Enable pretty printed json results when in dev mode.
	if !Production {
		gatewayHandler.Indent = "  "
	}
	graphqlURL := fmt.Sprintf("%s/graphql", config.server.URL)
	mux := http.NewServeMux()
	mux.Handle("/graphql", gatewayHandler)
	mux.Handle("/", graphiql.New(graphqlURL, true))
	config.server.Config.Handler = mux
	config.Log.Printf("GraphQL endpoint running at %s", graphqlURL)
	config.Log.Printf("GraphQL UI running at %s", config.server.URL)
	return nil
}

func (config *Config) postProcess() {
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
}
