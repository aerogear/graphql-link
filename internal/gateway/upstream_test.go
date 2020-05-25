package gateway_test

import (
	"fmt"
	"testing"

	"github.com/chirino/graphql-gw/internal/gateway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestUpstreamEncoding(t *testing.T) {

	expected := gateway.UpstreamWrapper{Upstream: &gateway.GraphQLUpstream{
		URL: "test",
	}}
	encoded, err := yaml.Marshal(expected)
	require.NoError(t, err)

	fmt.Println(string(encoded))

	actual := gateway.UpstreamWrapper{}
	err = yaml.Unmarshal(encoded, &actual)
	require.NoError(t, err)

	assert.Equal(t, actual, expected)
}
