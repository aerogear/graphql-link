package gateway_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql-gw/internal/gateway"
	"github.com/chirino/graphql-gw/internal/gateway/examples/characters"
	"github.com/chirino/graphql/httpgql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMountNamedFieldWithVariableNames(t *testing.T) {
	charactersEngine := characters.New()
	charactersServer := httptest.NewServer(&httpgql.Handler{ServeGraphQLStream: charactersEngine.ServeGraphQLStream})
	defer charactersServer.Close()

	engine, err := gateway.New(gateway.Config{
		Upstreams: map[string]gateway.UpstreamWrapper{
			"characters": {
				Upstream: &gateway.GraphQLUpstream{
					URL:    charactersServer.URL,
					Suffix: "_t1",
				},
			},
		},
		Types: []gateway.TypeConfig{
			{
				Name: `Query`,
				Actions: []gateway.ActionWrapper{
					{
						Action: &gateway.Mount{
							Field:    "mysearch",
							Upstream: "characters",
							Query: `query($text: String!) {
                           	   search(name:$text) 
                        	}`,
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, `type Character_t1 {
  id:ID!
  likes:Int!
  name:Name_t1
}
type Mutation {
}
type Name_t1 {
  first:String
  full:String
  last:String
}
type Query {
  mysearch(text:String!):Character_t1
}
type Subscription {
}
schema {
  mutation: Mutation
  query: Query
  subscription: Subscription
}
`, engine.Schema.String())

	server := httptest.NewServer(&httpgql.Handler{ServeGraphQLStream: engine.ServeGraphQLStream})
	defer server.Close()

	client := httpgql.NewClient(server.URL)
	res := client.ServeGraphQL(&graphql.Request{
		Query: `
{
	mysearch(text:"Rukia") { name { full }}
}`,
	})

	require.NoError(t, res.Error())
	assert.Equal(t, `{"mysearch":{"name":{"full":"Rukia Kuchiki"}}}`, string(res.Data))
}

func TestMountRootQueryOnNamedField(t *testing.T) {

	charactersEngine := characters.New()
	charactersServer := httptest.NewServer(&httpgql.Handler{ServeGraphQLStream: charactersEngine.ServeGraphQLStream})
	defer charactersServer.Close()

	gateway, err := gateway.New(gateway.Config{
		Upstreams: map[string]gateway.UpstreamWrapper{
			"characters": {
				Upstream: &gateway.GraphQLUpstream{
					URL:    charactersServer.URL,
					Suffix: "_t1",
				},
			},
		},
		Types: []gateway.TypeConfig{
			{
				Name: `Query`,
				Actions: []gateway.ActionWrapper{
					{
						Action: &gateway.Mount{
							Field:    "charactersQuery",
							Upstream: "characters",
							Query:    `query {}`,
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	ctx := context.Background()
	expected := map[string]interface{}{}
	err = charactersEngine.Exec(ctx, &expected, `
query  {
	characters {
	  id
	  name {
		first
		last
		full
	  }
	}
}`)
	require.NoError(t, err)

	actual := map[string]interface{}{}
	err = gateway.Exec(ctx, &actual, `
query  {
  charactersQuery {
    characters {
      id
      name {
        first
        last
        full
      }
    }
  }
}`)
	require.NoError(t, err)
	assert.Equal(t, expected, actual["charactersQuery"])

}

func TestMountAllFieldsOnRootQuery(t *testing.T) {

	charactersEngine := characters.New()
	charactersServer := httptest.NewServer(&httpgql.Handler{ServeGraphQLStream: charactersEngine.ServeGraphQLStream})
	defer charactersServer.Close()

	gateway, err := gateway.New(gateway.Config{
		Upstreams: map[string]gateway.UpstreamWrapper{
			"characters": {
				Upstream: &gateway.GraphQLUpstream{
					URL:    charactersServer.URL,
					Suffix: "_t1",
				},
			},
		},
		Types: []gateway.TypeConfig{
			{
				Name: `Query`,
				Actions: []gateway.ActionWrapper{
					{
						Action: &gateway.Mount{
							Upstream: "characters",
							Query:    `query {}`,
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, `type Character_t1 {
  id:ID!
  likes:Int!
  name:Name_t1
}
type Mutation {
}
type Name_t1 {
  first:String
  full:String
  last:String
}
type Query {
  characters:[Character_t1!]!
  search(name:String!):Character_t1
}
type Subscription {
}
schema {
  mutation: Mutation
  query: Query
  subscription: Subscription
}
`, gateway.Schema.String())

	require.NoError(t, err)
	ctx := context.Background()
	expected := map[string]interface{}{}
	err = charactersEngine.Exec(ctx, &expected, `
query  {
	characters {
	  id
	  name {
		first
		last
		full
	  }
	}
}`)
	require.NoError(t, err)

	actual := map[string]interface{}{}
	err = gateway.Exec(ctx, &actual, `
query  {
    characters {
      id
      name {
        first
        last
        full
      }
    }
}`)

	require.NoError(t, err)
	assert.Equal(t, expected, actual)

}

func TestMountNamedFieldWithArguments(t *testing.T) {

	charactersEngine := characters.New()
	charactersServer := httptest.NewServer(&httpgql.Handler{ServeGraphQLStream: charactersEngine.ServeGraphQLStream})
	defer charactersServer.Close()

	engine, err := gateway.New(gateway.Config{
		Upstreams: map[string]gateway.UpstreamWrapper{
			"characters": {
				Upstream: &gateway.GraphQLUpstream{
					URL:    charactersServer.URL,
					Suffix: "_t1",
				},
			},
		},
		Types: []gateway.TypeConfig{
			{
				Name: `Query`,
				Actions: []gateway.ActionWrapper{
					{
						Action: &gateway.Mount{
							Field:    "mysearch",
							Upstream: "characters",
							Query: `query {
                           search
                        }`,
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	server := httptest.NewServer(&httpgql.Handler{ServeGraphQLStream: engine.ServeGraphQLStream})
	defer server.Close()

	client := httpgql.NewClient(server.URL)
	res := client.ServeGraphQL(&graphql.Request{
		Query: `
{
	mysearch(name:"Rukia") { name { full }}
}`,
	})

	require.NoError(t, res.Error())
	assert.Equal(t, `{"mysearch":{"name":{"full":"Rukia Kuchiki"}}}`, string(res.Data))
}

func TestFieldAliases(t *testing.T) {

	charactersEngine := characters.New()
	charactersServer := httptest.NewServer(&httpgql.Handler{ServeGraphQLStream: charactersEngine.ServeGraphQLStream})
	defer charactersServer.Close()

	engine, err := gateway.New(gateway.Config{
		Upstreams: map[string]gateway.UpstreamWrapper{
			"characters": {
				Upstream: &gateway.GraphQLUpstream{
					URL:    charactersServer.URL,
					Suffix: "_t1",
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
	require.NoError(t, err)

	server := httptest.NewServer(&httpgql.Handler{ServeGraphQLStream: engine.ServeGraphQLStream})
	defer server.Close()

	client := httpgql.NewClient(server.URL)
	res := client.ServeGraphQL(&graphql.Request{
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
}

func TestSubscription(t *testing.T) {

	charactersEngine := characters.New()
	charactersServer := httptest.NewServer(&httpgql.Handler{ServeGraphQLStream: charactersEngine.ServeGraphQLStream})
	defer charactersServer.Close()

	engine, err := gateway.New(gateway.Config{
		Upstreams: map[string]gateway.UpstreamWrapper{
			"characters": {
				Upstream: &gateway.GraphQLUpstream{
					URL:    charactersServer.URL,
					Suffix: "_t1",
				},
			},
		},
		Types: []gateway.TypeConfig{
			{
				Name: `Subscription`,
				Actions: []gateway.ActionWrapper{
					{
						Action: &gateway.Mount{
							Upstream: "characters",
							Query:    `subscription {}`,
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	server := httptest.NewServer(&httpgql.Handler{ServeGraphQLStream: engine.ServeGraphQLStream})
	defer server.Close()

	client := httpgql.NewClient(server.URL)
	ctx, cancel := context.WithCancel(context.Background())
	resStream := client.ServeGraphQLStream(&graphql.Request{
		Context: ctx,
		Query: `
subscription {
	character(id:"1") { id, name { full }, likes }
}`,
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

}
