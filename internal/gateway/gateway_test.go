package gateway_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql-gw/internal/gateway"
	"github.com/chirino/graphql-gw/internal/gateway/examples/characters"
	"github.com/chirino/graphql-gw/internal/gateway/examples/starwars_characters"
	"github.com/chirino/graphql/httpgql"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ctx = context.Background()

func RunWithStarwarsGW(t *testing.T, c string, run func(gateway, client *httpgql.Client)) {
	upstreamEngine := starwars_characters.New()
	upstreamServer := httptest.NewServer(&httpgql.Handler{ServeGraphQLStream: upstreamEngine.ServeGraphQLStream})
	defer upstreamServer.Close()

	var config gateway.Config
	err := yaml.Unmarshal([]byte(c), &config)
	require.NoError(t, err)

	if config.Upstreams == nil {
		config.Upstreams = map[string]gateway.UpstreamWrapper{}
	}
	config.Upstreams["starwars"] = gateway.UpstreamWrapper{Upstream: &gateway.GraphQLUpstream{
		URL: upstreamServer.URL,
	}}
	engine, err := gateway.New(config)
	require.NoError(t, err)

	server := httptest.NewServer(&httpgql.Handler{ServeGraphQLStream: engine.ServeGraphQLStream})
	defer server.Close()

	run(httpgql.NewClient(server.URL), httpgql.NewClient(upstreamServer.URL))
}

func RunWithCharacterGW(t *testing.T, c string, run func(gateway, client *httpgql.Client)) {
	upstreamEngine := characters.New()
	upstreamServer := httptest.NewServer(&httpgql.Handler{ServeGraphQLStream: upstreamEngine.ServeGraphQLStream})
	defer upstreamServer.Close()

	var config gateway.Config
	err := yaml.Unmarshal([]byte(c), &config)
	require.NoError(t, err)

	if config.Upstreams == nil {
		config.Upstreams = map[string]gateway.UpstreamWrapper{}
	}
	config.Upstreams["characters"] = gateway.UpstreamWrapper{Upstream: &gateway.GraphQLUpstream{
		URL:    upstreamServer.URL,
		Suffix: "_t1",
	}}
	engine, err := gateway.New(config)
	require.NoError(t, err)

	server := httptest.NewServer(&httpgql.Handler{ServeGraphQLStream: engine.ServeGraphQLStream})
	defer server.Close()

	run(httpgql.NewClient(server.URL), httpgql.NewClient(upstreamServer.URL))
}

func TestRootTypeNames(t *testing.T) {

	// Verify default config
	config := gateway.Config{}
	engine, err := gateway.New(config)
	require.NoError(t, err)
	assert.Equal(t, `type Mutation {
}
type Query {
}
type Subscription {
}
schema {
  mutation: Mutation
  query: Query
  subscription: Subscription
}
`, engine.Schema.String())

	config = gateway.Config{
		Schema: gateway.SchemaConfig{
			Query:        "Q",
			Mutation:     "M",
			Subscription: "S",
		},
	}
	engine, err = gateway.New(config)
	require.NoError(t, err)
	assert.Equal(t, `type M {
}
type Q {
}
type S {
}
schema {
  mutation: M
  query: Q
  subscription: S
}
`, engine.Schema.String())
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

func TestComplexQuery(t *testing.T) {
	RunWithStarwarsGW(t, `
      types:
        - name: Query
          actions:
            - type: mount
              upstream: starwars
              query: query {}
`,
		func(gateway, characters *httpgql.Client) {
			res := gateway.ServeGraphQL(&graphql.Request{
				Variables: json.RawMessage(`{"episode":"JEDI", "withoutFriends": true, "withFriends": false}`),
				Query:     starwars_characters.ComplexStarWarsCharacterQuery,
			})
			require.NoError(t, res.Error())

			expected := `{
  "hero": {
    "id": "2001",
    "name": "R2-D2",
    "friends": [
      {
        "name": "Luke Skywalker"
      },
      {
        "name": "Han Solo"
      },
      {
        "name": "Leia Organa"
      }
    ]
  },
  "empireHerhero": {
    "name": "Luke Skywalker"
  },
  "jediHero": {
    "name": "R2-D2"
  },
  "human": {
    "name": "Luke Skywalker",
    "height": 1.72
  },
  "leftComparison": {
    "name": "Luke Skywalker",
    "appearsIn": [
      "NEWHOPE",
      "EMPIRE",
      "JEDI"
    ],
    "friends": [
      {
        "name": "Han Solo"
      },
      {
        "name": "Leia Organa"
      },
      {
        "name": "C-3PO"
      },
      {
        "name": "R2-D2"
      }
    ],
    "height": 1.72
  },
  "rightComparison": {
    "name": "R2-D2",
    "appearsIn": [
      "NEWHOPE",
      "EMPIRE",
      "JEDI"
    ],
    "friends": [
      {
        "name": "Luke Skywalker"
      },
      {
        "name": "Han Solo"
      },
      {
        "name": "Leia Organa"
      }
    ]
  },
  "heroNameAndFriends": {
    "name": "R2-D2"
  },
  "heroSkip": {
    "name": "R2-D2"
  },
  "heroInclude": {
    "name": "R2-D2"
  },
  "inlineFragments": {
    "name": "R2-D2",
    "primaryFunction": "Astromech"
  },
  "search": [
    {
      "__typename": "Human",
      "name": "Han Solo"
    },
    {
      "__typename": "Human",
      "name": "Leia Organa"
    }
  ],
  "heroConnections": {
    "name": "R2-D2",
    "friendsConnection": {
      "totalCount": 3,
      "pageInfo": {
        "startCursor": "Y3Vyc29yMQ==",
        "endCursor": "Y3Vyc29yMw==",
        "hasNextPage": false
      },
      "edges": [
        {
          "cursor": "Y3Vyc29yMQ==",
          "node": {
            "name": "Luke Skywalker"
          }
        },
        {
          "cursor": "Y3Vyc29yMg==",
          "node": {
            "name": "Han Solo"
          }
        },
        {
          "cursor": "Y3Vyc29yMw==",
          "node": {
            "name": "Leia Organa"
          }
        }
      ]
    }
  },
  "__schema": {
    "types": [
      {
        "name": "Boolean"
      },
      {
        "name": "Character"
      },
      {
        "name": "Droid"
      },
      {
        "name": "Episode"
      },
      {
        "name": "Float"
      },
      {
        "name": "FriendsConnection"
      },
      {
        "name": "FriendsEdge"
      },
      {
        "name": "Human"
      },
      {
        "name": "ID"
      },
      {
        "name": "Int"
      },
      {
        "name": "LengthUnit"
      },
      {
        "name": "Mutation"
      },
      {
        "name": "PageInfo"
      },
      {
        "name": "Query"
      },
      {
        "name": "SearchResult"
      },
      {
        "name": "Starship"
      },
      {
        "name": "String"
      },
      {
        "name": "Subscription"
      },
      {
        "name": "__Directive"
      },
      {
        "name": "__DirectiveLocation"
      },
      {
        "name": "__EnumValue"
      },
      {
        "name": "__Field"
      },
      {
        "name": "__InputValue"
      },
      {
        "name": "__Schema"
      },
      {
        "name": "__Type"
      },
      {
        "name": "__TypeKind"
      }
    ]
  },
  "__type": {
    "name": "Droid",
    "fields": [
      {
        "name": "id",
        "args": [],
        "type": {
          "name": null,
          "kind": "NON_NULL"
        }
      },
      {
        "name": "name",
        "args": [],
        "type": {
          "name": null,
          "kind": "NON_NULL"
        }
      },
      {
        "name": "friends",
        "args": [],
        "type": {
          "name": null,
          "kind": "LIST"
        }
      },
      {
        "name": "friendsConnection",
        "args": [
          {
            "name": "first",
            "type": {
              "name": "Int"
            },
            "defaultValue": null
          },
          {
            "name": "after",
            "type": {
              "name": "ID"
            },
            "defaultValue": null
          }
        ],
        "type": {
          "name": null,
          "kind": "NON_NULL"
        }
      },
      {
        "name": "appearsIn",
        "args": [],
        "type": {
          "name": null,
          "kind": "NON_NULL"
        }
      },
      {
        "name": "primaryFunction",
        "args": [],
        "type": {
          "name": "String",
          "kind": "SCALAR"
        }
      }
    ]
  }
}`
			//t.Skip("the returned data is still not exactly right...")
			got, err := json.MarshalIndent(res.Data, "", "  ")
			require.NoError(t, err)
			assert.Equal(t, expected, string(got))
		})

}
