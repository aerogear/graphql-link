package gateway_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql-gw/internal/gateway"
	"github.com/chirino/graphql-gw/internal/gateway/examples/characters"
	"github.com/chirino/graphql/httpgql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testHandler struct {
	requestCounter int64
	lastRequest    *http.Request
	*httpgql.Handler
	mu sync.Mutex
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	h.lastRequest = r
	h.requestCounter++
	h.mu.Unlock()
	h.Handler.ServeHTTP(w, r)
}

func (h *testHandler) GetLastRequest() (r *http.Request) {
	h.mu.Lock()
	r = h.lastRequest
	h.mu.Unlock()
	return
}

func (h *testHandler) GetRequestCounter() (r int64) {
	h.mu.Lock()
	r = h.requestCounter
	h.mu.Unlock()
	return
}

func TestProxyHeaders(t *testing.T) {

	charactersEngine := characters.New()

	h := &testHandler{Handler: &httpgql.Handler{ServeGraphQLStream: charactersEngine.ServeGraphQLStream}}
	charactersServer := httptest.NewServer(h)
	defer charactersServer.Close()

	config := createCharactersPassthroughConfig()
	config.Upstreams["characters"].Upstream.(*gateway.GraphQLUpstream).URL = charactersServer.URL
	gw, err := gateway.New(config)
	require.NoError(t, err)
	gatewayServer := httptest.NewServer(gateway.CreateHttpHandler(gw.ServeGraphQLStream))
	defer gatewayServer.Close()

	client := httpgql.NewClient(gatewayServer.URL)

	actual := map[string]interface{}{}
	graphql.Exec(client.ServeGraphQL, context.Background(), &actual, `
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

	request := h.GetLastRequest()
	value := request.Header.Get("X-Forwarded-For")
	assert.Equal(t, `127.0.0.1`, value)

	value = request.Header.Get("X-Forwarded-Host")
	assert.Equal(t, ``, value)

	value = request.Header.Get("X-Forwarded-Proto")
	assert.Equal(t, `http`, value)
}

func TestCustomClientHeadersArePassedThrough(t *testing.T) {

	charactersEngine := characters.New()

	h := &testHandler{Handler: &httpgql.Handler{ServeGraphQLStream: charactersEngine.ServeGraphQLStream}}
	charactersServer := httptest.NewServer(h)
	defer charactersServer.Close()

	config := createCharactersPassthroughConfig()
	config.Upstreams["characters"].Upstream.(*gateway.GraphQLUpstream).URL = charactersServer.URL
	gw, err := gateway.New(config)
	require.NoError(t, err)
	gatewayServer := httptest.NewServer(gateway.CreateHttpHandler(gw.ServeGraphQLStream))
	defer gatewayServer.Close()

	client := httpgql.NewClient(gatewayServer.URL)
	client.RequestHeader.Set("Custom", "Hello World")

	actual := map[string]interface{}{}
	graphql.Exec(client.ServeGraphQL, context.Background(), &actual, `
query  {
    characters {
      id
    }
}`)

	request := h.GetLastRequest()
	value := request.Header.Get("Custom")
	assert.Equal(t, `Hello World`, value)
}

func TestSingleHopHeadersAreNotPassedThrough(t *testing.T) {

	charactersEngine := characters.New()

	h := &testHandler{Handler: &httpgql.Handler{ServeGraphQLStream: charactersEngine.ServeGraphQLStream}}
	charactersServer := httptest.NewServer(h)
	defer charactersServer.Close()

	config := createCharactersPassthroughConfig()
	config.Upstreams["characters"].Upstream.(*gateway.GraphQLUpstream).URL = charactersServer.URL
	gw, err := gateway.New(config)
	require.NoError(t, err)
	gatewayServer := httptest.NewServer(gateway.CreateHttpHandler(gw.ServeGraphQLStream))
	defer gatewayServer.Close()

	client := httpgql.NewClient(gatewayServer.URL)
	client.RequestHeader.Set("Proxy-Authenticate", "Hello World")

	actual := map[string]interface{}{}
	graphql.Exec(client.ServeGraphQL, context.Background(), &actual, `
query  {
    characters {
      id
    }
}`)

	request := h.GetLastRequest()
	value := request.Header.Get("Custom")
	assert.Equal(t, ``, value)
}

func createCharactersPassthroughConfig() gateway.Config {
	return gateway.Config{
		Upstreams: map[string]gateway.UpstreamWrapper{
			"characters": {
				Upstream: &gateway.GraphQLUpstream{
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
	}
}

func TestDataLoader(t *testing.T) {

	charactersEngine := characters.New()

	h := &testHandler{Handler: &httpgql.Handler{ServeGraphQLStream: charactersEngine.ServeGraphQLStream}}
	charactersServer := httptest.NewServer(h)
	defer charactersServer.Close()

	config := createCharactersPassthroughConfig()
	config.Upstreams["characters"].Upstream.(*gateway.GraphQLUpstream).URL = charactersServer.URL
	gw, err := gateway.New(config)
	require.NoError(t, err)

	// The first request is an introspection query.
	assert.Equal(t, int64(1), h.GetRequestCounter())

	actual := json.RawMessage{}
	graphql.Exec(gw.ServeGraphQL, context.Background(), &actual, `
query  {
    search(name:"Rukia") {
      id
    }
    characters {
      id
    }
}`)
	// Verify that both those root selection get aggregated into a single
	// query to the upstream server.
	assert.Equal(t, `{"search":{"id":"1"},"characters":[{"id":"1"},{"id":"2"},{"id":"3"},{"id":"3"},{"id":"3"},{"id":"3"}]}`, string(actual))
	assert.Equal(t, int64(2), h.GetRequestCounter())

}
