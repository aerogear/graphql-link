package gateway_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/aerogear/graphql-link/internal/gateway"
	"github.com/aerogear/graphql-link/internal/gateway/examples/characters"
	"github.com/chirino/graphql"
	"github.com/chirino/graphql/httpgql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
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

	h, charactersServer, gatewayServer, client := createTestServersWithTestHandler(t, createCharactersPassthroughConfig())
	defer charactersServer.Close()
	defer gatewayServer.Close()

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

	h, charactersServer, gatewayServer, client := createTestServersWithTestHandler(t, createCharactersPassthroughConfig())
	defer charactersServer.Close()
	defer gatewayServer.Close()

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

func TestDisabledCustomClientHeadersPassThrough(t *testing.T) {

	config := mustCreateConfig(`
upstreams:
  characters:
    suffix: _t1
    headers:
      disable-forwarding: true
types:
  - name: Query
    actions:
      - type: mount
        upstream: characters
        query: query {}
`)
	h, charactersServer, gatewayServer, client := createTestServersWithTestHandler(t, config)
	defer charactersServer.Close()
	defer gatewayServer.Close()

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
	assert.Equal(t, ``, value)
}

func TestUpstreamSetHeader(t *testing.T) {

	config := mustCreateConfig(`
upstreams:
  characters:
    suffix: _t1
    headers:
      set:
        - name: Custom
          value: Upstream Set
types:
  - name: Query
    actions:
      - type: mount
        upstream: characters
        query: query {}
`)
	h, charactersServer, gatewayServer, client := createTestServersWithTestHandler(t, config)
	defer charactersServer.Close()
	defer gatewayServer.Close()

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
	assert.Equal(t, `Upstream Set`, value)
}

func TestUpstreamRemoveHeader(t *testing.T) {

	config := mustCreateConfig(`
upstreams:
  characters:
    suffix: _t1
    headers:
      remove:
        - Custom 
types:
  - name: Query
    actions:
      - type: mount
        upstream: characters
        query: query {}
`)
	h, charactersServer, gatewayServer, client := createTestServersWithTestHandler(t, config)
	defer charactersServer.Close()
	defer gatewayServer.Close()

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
	assert.Equal(t, ``, value)
}

func createTestServersWithTestHandler(t *testing.T, config gateway.Config) (*testHandler, *httptest.Server, *httptest.Server, *httpgql.Client) {
	charactersEngine := characters.New()
	h := &testHandler{Handler: &httpgql.Handler{ServeGraphQLStream: charactersEngine.ServeGraphQLStream}}
	charactersServer := httptest.NewServer(h)

	config.Upstreams["characters"].Upstream.(*gateway.GraphQLUpstream).URL = charactersServer.URL
	gw, err := gateway.New(config)
	require.NoError(t, err)
	gatewayServer := httptest.NewServer(gateway.CreateHttpHandler(gw.ServeGraphQLStream))

	client := httpgql.NewClient(gatewayServer.URL)
	return h, charactersServer, gatewayServer, client
}

func TestSingleHopHeadersAreNotPassedThrough(t *testing.T) {

	h, charactersServer, gatewayServer, client := createTestServersWithTestHandler(t, createCharactersPassthroughConfig())
	defer charactersServer.Close()
	defer gatewayServer.Close()

	client.RequestHeader.Set("Proxy-Authenticate", "Hello World")
	actual := map[string]interface{}{}
	graphql.Exec(client.ServeGraphQL, context.Background(), &actual, `
query  {
    characters {
      id
    }
}`)

	request := h.GetLastRequest()
	value := request.Header.Get("Proxy-Authenticate")
	assert.Equal(t, ``, value)
}

func mustCreateConfig(gatewayConfig string) gateway.Config {
	var config gateway.Config
	err := yaml.Unmarshal([]byte(gatewayConfig), &config)
	if err != nil {
		panic(err)
	}
	return config
}

func createCharactersPassthroughConfig() gateway.Config {
	return mustCreateConfig(`
upstreams:
  characters:
    suffix: _t1

types:
  - name: Query
    actions:
      - type: mount
        upstream: characters
        query: query {}
`)
}

func TestDataLoader(t *testing.T) {

	h, charactersServer, gatewayServer, gw := createTestServersWithTestHandler(t, createCharactersPassthroughConfig())
	defer charactersServer.Close()
	defer gatewayServer.Close()

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
