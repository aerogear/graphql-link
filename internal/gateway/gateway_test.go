package gateway_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql-gw/internal/gateway"
	"github.com/chirino/graphql-gw/internal/gateway/examples/characters"
	"github.com/chirino/graphql/httpgql"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ctx = context.Background()

func RunWithCharacterGW(t *testing.T, c string, run func(gateway, characters *httpgql.Client)) {
	charactersEngine := characters.New()
	charactersServer := httptest.NewServer(&httpgql.Handler{ServeGraphQLStream: charactersEngine.ServeGraphQLStream})
	defer charactersServer.Close()

	var config gateway.Config
	err := yaml.Unmarshal([]byte(c), &config)
	require.NoError(t, err)

	if config.Upstreams == nil {
		config.Upstreams = map[string]gateway.UpstreamWrapper{}
	}
	config.Upstreams["characters"] = gateway.UpstreamWrapper{Upstream: &gateway.GraphQLUpstream{
		URL:    charactersServer.URL,
		Suffix: "_t1",
	}}
	engine, err := gateway.New(config)
	require.NoError(t, err)

	server := httptest.NewServer(&httpgql.Handler{ServeGraphQLStream: engine.ServeGraphQLStream})
	defer server.Close()

	run(httpgql.NewClient(server.URL), httpgql.NewClient(charactersServer.URL))
}

func TestFieldAliases(t *testing.T) {
	RunWithCharacterGW(t, `
      types:
        - name: Query
          actions:
            - type: mount
              field: characters
              upstream: characters
              query: query {}
            - type: mount
              field: rukiaId
              upstream: characters
              query: |
                query { search(name: "Rukia") { id } }
`,
		func(gateway, characters *httpgql.Client) {
			res := gateway.ServeGraphQL(&graphql.Request{
				Query: `
					query anilist {
					  y: rukiaId
					  z: characters {
						y:search(name: "Rukia") {
						  x: id
						}
					  }
					}`,
			})
			require.NoError(t, res.Error())
			assert.Equal(t, `{"y":"1","z":{"y":{"x":"1"}}}`, string(res.Data))
		})
}

func TestSubscription(t *testing.T) {
	RunWithCharacterGW(t, `
      types:
        - name: Subscription
          actions:
            - type: mount
              upstream: characters
              query: subscription {}`,
		func(gateway, characters *httpgql.Client) {

			ctx, cancel := context.WithCancel(ctx)
			resStream := gateway.ServeGraphQLStream(&graphql.Request{
				Context: ctx,
				Query:   `subscription { character(id:"1") { id, name { full }, likes } }`,
			})

			res := <-resStream
			require.NoError(t, res.Error())
			assert.Equal(t, `{"character":{"id":"1","name":{"full":"Rukia Kuchiki"},"likes":1}}`, string(res.Data))

			res = <-resStream
			require.NoError(t, res.Error())
			assert.Equal(t, `{"character":{"id":"1","name":{"full":"Rukia Kuchiki"},"likes":2}}`, string(res.Data))

			res = <-resStream
			require.NoError(t, res.Error())
			assert.Equal(t, `{"character":{"id":"1","name":{"full":"Rukia Kuchiki"},"likes":3}}`, string(res.Data))

			cancel()
		})
}

func TestMutationWithObjectInput(t *testing.T) {
	RunWithCharacterGW(t, `
      types:
        - name: Mutation
          actions:
            - type: mount
              upstream: characters
              query: mutation {}`,
		func(gateway, characters *httpgql.Client) {
			res := gateway.ServeGraphQL(&graphql.Request{
				Variables: json.RawMessage(`{"character":{"name":{"first":"Hiram", "last":"Chirino"}}}`),
				Query: `
					mutation($character:CharacterInput_t1!) {
						add(character:$character) {
							name { full }
						}
					}`,
			})
			require.NoError(t, res.Error())
			assert.Equal(t, `{"add":{"name":{"full":"Hiram Chirino"}}}`, string(res.Data))
		})

}
