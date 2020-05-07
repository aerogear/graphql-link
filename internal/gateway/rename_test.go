package gateway_test

import (
	"testing"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql/httpgql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenameField(t *testing.T) {
	RunWithCharacterGW(t, `
      types:
        - name: Query
          actions:
            - type: mount
              upstream: characters
              query: query {}
            - type: rename
              field: search
              to: find`,
		func(gateway, characters *httpgql.Client) {
			res := gateway.ServeGraphQL(&graphql.Request{
				Query: ` query { search(name: "Rukia") { x: id } }`,
			})
			require.Error(t, res.Error())

			// make sure the original field name is not there.
			res = gateway.ServeGraphQL(&graphql.Request{
				Query: ` query { find(name: "Rukia") { x: id } }`,
			})
			require.NoError(t, res.Error())
			assert.Equal(t, `{"find":{"x":"1"}}`, string(res.Data))
		})
}

func TestRenameType(t *testing.T) {
	RunWithCharacterGW(t, `
      types:
        - name: Query
          actions:
            - type: mount
              upstream: characters
              query: query {}
            - type: rename
              to: RenamedQuery`,
		func(gateway, characters *httpgql.Client) {
			res := gateway.ServeGraphQL(&graphql.Request{
				Query: ` query { __typename }`,
			})
			require.NoError(t, res.Error())
			assert.Equal(t, `{"__typename":"RenamedQuery"}`, string(res.Data))
		})
}
