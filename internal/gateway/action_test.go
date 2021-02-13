package gateway_test

import (
	"fmt"
	"testing"

	"github.com/aerogear/graphql-link/internal/gateway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestMarshalAction(t *testing.T) {

	expected := gateway.ActionWrapper{Action: &gateway.Link{
		Field: "test",
	}}
	encoded, err := yaml.Marshal(expected)
	require.NoError(t, err)

	fmt.Println(string(encoded))

	actual := gateway.ActionWrapper{}
	err = yaml.Unmarshal(encoded, &actual)
	require.NoError(t, err)

	assert.Equal(t, actual, expected)
}
