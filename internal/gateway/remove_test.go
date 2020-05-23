package gateway_test

import (
	"testing"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql/httpgql"
	"github.com/stretchr/testify/require"
)

func TestRemoveField(t *testing.T) {
	RunWithCharacterGW(t, `
      types:
        - name: Query
          actions:
            - type: mount
              upstream: characters
              query: query {}
            - type: remove
              field: search`,
		func(gateway, characters *httpgql.Client) {
			res := gateway.ServeGraphQL(&graphql.Request{
				Query: ` query { search(name: "Rukia") { x: id } }`,
			})
			require.Error(t, res.Error())
		})
}
