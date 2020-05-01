package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql-gw/internal/gateway/examples/characters"
	"github.com/chirino/graphql-gw/internal/gateway/examples/shows"
	"github.com/chirino/graphql-gw/internal/gateway/examples/starwars_characters"
	"github.com/chirino/graphql/graphiql"
	"github.com/chirino/graphql/relay"
)

func main() {
	charactersEngine := characters.New()
	charactersServer := StartupServer("0.0.0.0", 8081, charactersEngine)
	defer charactersServer.Close()

	showsEngine := shows.New()
	showsServer := StartupServer("0.0.0.0", 8082, showsEngine)
	defer showsServer.Close()

	starwars_charactersEngine := starwars_characters.New()
	starwars_charactersServer := StartupServer("0.0.0.0", 8083, starwars_charactersEngine)
	defer starwars_charactersServer.Close()

	for {
		time.Sleep(time.Hour)
	}
}

func StartupServer(host string, port uint16, engine *graphql.Engine) *httptest.Server {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		switch host {
		case "localhost":
			fallthrough
		case "127.0.0.1":
			host = "[::1]"
			if l, err = net.Listen("tcp6", fmt.Sprintf("%s:%d", host, port)); err != nil {
				panic(fmt.Sprintf("httptest: failed to listen on a port: %v", err))
			}
		default:
			panic(fmt.Sprintf("httptest: failed to listen on a port: %v", err))
		}
	}

	mux := http.NewServeMux()
	mux.Handle("/graphql", &relay.Handler{ServeGraphQLStream: engine.ServeGraphQLStream})
	endpoint := fmt.Sprintf("http://%s:%d", host, port)
	mux.Handle("/", graphiql.New(endpoint+"/graphql", false))
	ts := &httptest.Server{
		Listener: l,
		Config:   &http.Server{Handler: mux},
	}
	log.Printf("GraphQL endpoint running at %s/graphql", endpoint)
	log.Printf("GraphQL UI running at %s", endpoint)
	ts.Start()
	return ts
}
