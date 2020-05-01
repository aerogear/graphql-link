package gateway_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql-gw/internal/gateway"
	"github.com/chirino/graphql-gw/internal/gateway/examples/characters"
	"github.com/chirino/graphql/relay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testHandler struct {
	lastRequest *http.Request
	*relay.Handler
	mu sync.Mutex
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	h.lastRequest = r
	h.mu.Unlock()
	h.Handler.ServeHTTP(w, r)
}

func (h *testHandler) GetLastRequest() (r *http.Request) {
	h.mu.Lock()
	r = h.lastRequest
	h.mu.Unlock()
	return
}

func TestProxyHeaders(t *testing.T) {

	charactersEngine := characters.New()

	h := &testHandler{Handler: &relay.Handler{ServeGraphQLStream: charactersEngine.ServeGraphQLStream}}
	charactersServer := httptest.NewServer(h)
	defer charactersServer.Close()

	gw, err := gateway.New(gateway.Config{
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
	gatewayServer := httptest.NewServer(gateway.CreateHttpHandler(gw.ServeGraphQLStream))
	defer gatewayServer.Close()

	client := relay.NewClient(gatewayServer.URL)

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
