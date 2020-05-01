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

	charactersEngine := characters.New()
	charactersServer := httptest.NewServer(&relay.Handler{ServeGraphQLStream: charactersEngine.ServeGraphQLStream})
	defer charactersServer.Close()

	showsEngine := shows.New()
	showsServer := httptest.NewServer(&relay.Handler{ServeGraphQLStream: showsEngine.ServeGraphQLStream})
	defer showsServer.Close()

	engine, err := gateway.New(gateway.Config{
		Upstreams: map[string]gateway.UpstreamWrapper{
			"characters": {
				Upstream: &gateway.GraphQLUpstream{
					URL:    charactersServer.URL,
					Suffix: "_t1",
				},
			},
			"shows": {
				Upstream: &gateway.GraphQLUpstream{
					URL:    showsServer.URL,
					Suffix: "_t2",
				},
			},
		},
		Types: []gateway.TypeConfig{
			{
				Name: `Query`,
				Actions: []gateway.ActionWrapper{
					{
						Action: &gateway.Mount{
							Field:    "characters",
							Upstream: "characters",
							Query:    `query {}`,
						},
					},
					{
						Action: &gateway.Mount{
							Field:    "shows",
							Upstream: "shows",
							Query:    `query {}`,
						},
					},
					{
						Action: &gateway.Mount{
							Field:    "rukiaId",
							Upstream: "characters",
							Query: `query {
   									search(name: "Rukia") {
										id
									}
								}`,
						},
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
