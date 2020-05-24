package gateway_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/chirino/graphql-gw/internal/gateway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalAction(t *testing.T) {

	expected := gateway.ActionWrapper{Action: &gateway.Link{
		Field: "test",
	}}
	encoded, err := json.Marshal(expected)
	require.NoError(t, err)

	fmt.Println(string(encoded))

	actual := gateway.ActionWrapper{}
	err = json.Unmarshal(encoded, &actual)
	require.NoError(t, err)

	assert.Equal(t, actual, expected)
}
