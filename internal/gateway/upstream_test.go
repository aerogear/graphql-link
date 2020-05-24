package gateway_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/chirino/graphql-gw/internal/gateway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpstreamEncoding(t *testing.T) {

	expected := gateway.UpstreamWrapper{Upstream: &gateway.GraphQLUpstream{
		URL: "test",
	}}
	encoded, err := json.Marshal(expected)
	require.NoError(t, err)

	fmt.Println(string(encoded))

	actual := gateway.UpstreamWrapper{}
	err = json.Unmarshal(encoded, &actual)
	require.NoError(t, err)

	assert.Equal(t, actual, expected)
}
