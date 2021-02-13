package gateway_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"testing"

	"github.com/chirino/graphql"
	"github.com/aerogear/graphql-link/internal/gateway"
	gw "github.com/aerogear/graphql-link/internal/gateway/policyagent/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type mockPolicyAgentServer struct{}

func (s *mockPolicyAgentServer) Check(ctx context.Context, req *gw.CheckRequest) (*gw.CheckResponse, error) {
	resp := gw.CheckResponse{}
	for i, f := range req.Graphql.GetFields() {
		if (i % 2) == 1 {
			resp.Fields = append(resp.Fields, &gw.GraphQLFieldResponse{
				Path:  f.Path,
				Error: "You are not allowed to access odd fields",
			})
		} else {
			//resp.Fields = append(resp.Fields, &gw.GraphQLFieldResponse{
			//	Path: f.Path,
			//	//SetHeaders: []*gw.Header{
			//	//	{Name: "Upstream-Key", Value: "foo"},
			//	//},
			//})
		}
	}
	return &resp, nil
}

func TestPolicyAgent(t *testing.T) {

	grpcServer, grpcAddress := startGRPCServer(t, &mockPolicyAgentServer{})
	defer grpcServer.Stop()

	config := createCharactersPassthroughWithPolicyAgentConfig()
	config.PolicyAgent.Address = grpcAddress

	_, charactersServer, gatewayServer, client := createTestServersWithTestHandler(t, config)
	defer charactersServer.Close()
	defer gatewayServer.Close()

	resp := client.ServeGraphQL(&graphql.Request{
		Query: `
query  {
    characters {
      id
      name {
        first
        last
        full
      }
    }
}`})
	data, err := json.Marshal(resp)
	require.NoError(t, err)
	fmt.Println(string(data))
	assert.Equal(t,
		`{"data":{"characters":[{"name":{"last":"Kuchiki"}},{"name":{"last":"Kurosaki"}},{"name":{"last":"Inoue"}},{"name":{"last":""}},{"name":{"last":""}},{"name":{"last":"Schuberg"}}]},`+
			`"errors":[{"message":"You are not allowed to access odd fields","path":["characters","id"]},{"message":"You are not allowed to access odd fields","path":["characters","name","first"]},{"message":"You are not allowed to access odd fields","path":["characters","name","full"]},{"message":"You are not allowed to access odd fields","path":["characters","id"]},{"message":"You are not allowed to access odd fields","path":["characters","name","first"]},{"message":"You are not allowed to access odd fields","path":["characters","name","full"]},{"message":"You are not allowed to access odd fields","path":["characters","id"]},{"message":"You are not allowed to access odd fields","path":["characters","name","first"]},{"message":"You are not allowed to access odd fields","path":["characters","name","full"]},{"message":"You are not allowed to access odd fields","path":["characters","id"]},{"message":"You are not allowed to access odd fields","path":["characters","name","first"]},{"message":"You are not allowed to access odd fields","path":["characters","name","full"]},{"message":"You are not allowed to access odd fields","path":["characters","id"]},{"message":"You are not allowed to access odd fields","path":["characters","name","first"]},{"message":"You are not allowed to access odd fields","path":["characters","name","full"]},{"message":"You are not allowed to access odd fields","path":["characters","id"]},{"message":"You are not allowed to access odd fields","path":["characters","name","first"]},{"message":"You are not allowed to access odd fields","path":["characters","name","full"]}]}`,
		string(data))
}

func startGRPCServer(t *testing.T, server gw.PolicyAgentServer) (*grpc.Server, string) {
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	grpcServer := grpc.NewServer()
	gw.RegisterPolicyAgentServer(grpcServer, server)
	go grpcServer.Serve(lis)
	add := lis.Addr().String()
	return grpcServer, add
}

func createCharactersPassthroughWithPolicyAgentConfig() gateway.Config {
	return mustCreateConfig(`
upstreams:
  characters:
    suffix: _t1
policy-agent:
  insecure-client: true
types:
  - name: Query
    actions:
      - type: mount
        upstream: characters
        query: query {}
`)
}
