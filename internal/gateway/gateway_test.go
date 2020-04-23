package gateway_test

import (
	"fmt"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql-gw/internal/gateway"
	"github.com/chirino/graphql/relay"
	"github.com/chirino/graphql/resolvers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGateway(t *testing.T) {
	helloServer := startHelloGraphQL()
	defer helloServer.Close()

	engine, err := gateway.NewEngine(gateway.Config{
		Endpoints: map[string]gateway.EndpointInfo{
			"hello": {
				URL: helloServer.URL,
			},
		},
		Query: map[string]gateway.Field{
			"hi": {
				Description: "hi: provided by the hello service",
				Endpoint:    "hello",
				Query: `
query($tok: String!, $firstName:String!) {
	login(token:$tok) {
		hello(name:$firstName)
	}
}`,
			},
		},
		Mutation: map[string]gateway.Field{},
	})
	require.NoError(t, err)

	assert.Equal(t, `type Mutation {
}
type Query {
  "hi: provided by the hello service"
  hi(firstName:String!, tok:String!):String
}
schema {
  mutation: Mutation
  query: Query
}
`, engine.Schema.String())

	server := httptest.NewServer(&relay.Handler{Engine: engine})
	defer server.Close()

	client := relay.NewClient(server.URL)
	res := client.ServeGraphQL(&graphql.EngineRequest{
		Query: `
{
	hi(tok:"03D5FDA", firstName:"Hiram")
}`,
	})

	assert.NoError(t, res.Error())
	assert.Equal(t, `{"hi":"(03D5FDA): Hello Hiram"}`, string(res.Data))

}

func startHelloGraphQL() *httptest.Server {

	engine := graphql.New()
	engine.Schema.Parse(`
        schema {
            query: Query
        }
        type Query { 
            schema:String
            login(token:String!):LoggedIn
        }
        type LoggedIn { 
            hello(name:String!):String
        }
    `)

	engine.Resolver = resolvers.List(engine.Resolver, resolvers.Func(func(request *resolvers.ResolveRequest, next resolvers.Resolution) resolvers.Resolution {
		// use the built in resolver if it's available.
		if next != nil {
			return next
		}

		if request.Field.Name == "login" {
			return func() (value reflect.Value, err error) {
				return reflect.ValueOf(request.Args["token"]), nil
			}
		}
		if request.Field.Name == "hello" {
			return func() (value reflect.Value, err error) {
				rc := fmt.Sprintf("(%s): Hello %s", request.Parent.Interface(), request.Args["name"])
				return reflect.ValueOf(rc), nil
			}
		}
		return nil
	}))
	return httptest.NewServer(&relay.Handler{Engine: engine})
}
