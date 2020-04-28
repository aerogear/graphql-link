package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/chirino/graphql-gw/internal/gateway"
	"github.com/chirino/graphql-gw/internal/gateway/examples/characters"
	"github.com/chirino/graphql-gw/internal/gateway/examples/shows"
	"github.com/chirino/graphql/graphiql"
	"github.com/chirino/graphql/relay"
)

func main() {
	host := "localhost"
	port := "8080"

	chractersEngine := characters.New()
	chractersServer := httptest.NewServer(&relay.Handler{ServeGraphQLStream: chractersEngine.ServeGraphQLStream})
	defer chractersServer.Close()

	showsEngine := shows.New()
	showsServer := httptest.NewServer(&relay.Handler{ServeGraphQLStream: showsEngine.ServeGraphQLStream})
	defer showsServer.Close()

	engine, err := gateway.New(gateway.Config{
		Endpoints: map[string]gateway.EndpointInfo{
			"chracters": {
				URL:    chractersServer.URL,
				Suffix: "_t1",
			},
			"shows": {
				URL:    showsServer.URL,
				Suffix: "_t2",
			},
		},
		Types: []gateway.TypeConfig{
			{
				Name: `Query`,
				Fields: []gateway.Field{
					{
						Name:     "chracters",
						Endpoint: "chracters",
						Query:    `query {}`,
					},
					{
						Name:     "shows",
						Endpoint: "shows",
						Query:    `query {}`,
					},
				},
			},
		},
	})

	if err != nil {
		panic(err)
	}

	endpoint := fmt.Sprintf("http://%s:%s", host, port)
	http.Handle("/graphql", &relay.Handler{ServeGraphQLStream: engine.ServeGraphQLStream})
	log.Printf("GraphQL endpoint running at %s/graphql", endpoint)
	http.Handle("/", graphiql.New(endpoint+"/graphql", false))
	log.Printf("GraphQL UI running at %s", endpoint)

	log.Fatalf("%+v", http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), nil))

}
