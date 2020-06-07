package serve

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/chirino/graphql-gw/internal/cmd/config"
	"github.com/chirino/graphql-gw/internal/cmd/root"
	"github.com/chirino/graphql-gw/internal/gateway"
	"github.com/chirino/graphql-gw/internal/gateway/admin"
	"github.com/chirino/graphql-gw/internal/gateway/admin/assets"
	"github.com/chirino/graphql/graphiql"
	"github.com/chirino/graphql/httpgql"
	"github.com/fsnotify/fsnotify"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:               "serve",
		Short:             "Runs the gateway service",
		Run:               run,
		PersistentPreRunE: config.PreRunLoad,
	}
	Production = false
)

func init() {
	Command.Flags().StringVar(&config.File, "config", "graphql-gw.yaml", "path to the config file to load")
	Command.Flags().BoolVar(&Production, "production", false, "when true, the server will not download and store schemas from remote graphql endpoints.")
	root.Command.AddCommand(Command)
}

func run(_ *cobra.Command, _ []string) {
	lastConfig := *config.Value
	lastConfig.Log = gateway.TimestampedLog
	log := lastConfig.Log

	err := startServer(&lastConfig)
	if err != nil {
		log.Fatalf("could not start the sever: "+root.Verbosity, err)
	}

	restartMu := sync.Mutex{}
	restart := func() {
		restartMu.Lock()
		defer restartMu.Unlock()

		nextConfig := lastConfig
		err := config.Load(&nextConfig)
		if err != nil {
			log.Fatalf("error reading configuration file: "+root.Verbosity, err)
		}
		nextConfig.Log = lastConfig.Log

		// restart the server on a port change...
		if lastConfig.Listen != nextConfig.Listen {
			lastConfig.Server.Close()

			err = startServer(&nextConfig)
			if err != nil {
				log.Fatalf("could not start the sever: "+root.Verbosity, err)
			}
		} else {
			// Just remount a new engine on to the server...
			postProcess(&nextConfig)
			err = mountGatewayOnHttpServer(&nextConfig)
			if err != nil {
				log.Fatalf("could not start the sever: "+root.Verbosity, err)
			}
		}

		lastConfig = nextConfig
	}

	if !Production {
		watchFile(config.File, func(in fsnotify.Event) {
			log.Println("restarting due to configuration change:", in.Name)
			restart()
		})

		go func() {
			for {
				changed, err := gateway.HaveUpstreamSchemaChanged(lastConfig.Config)
				if err == nil && changed {
					log.Println("restarting due to an upstream schema change")
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
			log.Println("shutting down due to SIGINT signal")
			lastConfig.Server.Close()
			os.Exit(0)

		case syscall.SIGTERM:
			log.Println("shutting down due to SIGTERM signal")
			lastConfig.Server.Close()
			os.Exit(0)

		case syscall.SIGHUP:
			log.Println("restarting due to SIGHUP signal")
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

func startServer(config *config.Config) error {
	postProcess(config)
	server, err := gateway.StartHttpListener(config.Listen, http.NewServeMux())
	if err != nil {
		return err
	}
	config.Server = server
	return mountGatewayOnHttpServer(config)
}

func mountGatewayOnHttpServer(c *config.Config) error {
	engine, err := gateway.New(c.Config)
	if err != nil {
		return err
	}
	gatewayHandler := gateway.CreateHttpHandler(engine.ServeGraphQLStream).(*httpgql.Handler)
	// Enable pretty printed json results when in dev mode.
	if !Production {
		gatewayHandler.Indent = "  "
	}
	graphqlURL := fmt.Sprintf("%s/graphql", c.Server.URL)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	if !Production {
		r.Mount("/", http.FileServer(assets.FileSystem))
		r.Mount("/admin", admin.CreateHttpHandler())
	}
	r.Handle("/graphql", gatewayHandler)
	r.Handle("/graphiql", graphiql.New(graphqlURL, true))
	c.Server.Config.Handler = r
	c.Config.Log.Printf("GraphQL endpoint is running at %s", graphqlURL)
	if Production {
		c.Config.Log.Printf("Gateway GraphQL IDE is running at %s/graphiql", c.Server.URL)
	} else {
		c.Config.Log.Printf("Gateway Admin UI and GraphQL IDE is running at %s", c.Server.URL)
	}
	return nil
}

func postProcess(config *config.Config) {
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
