package gateway_test

import (
	"testing"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql/httpgql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMountNamedFieldWithVariableNames(t *testing.T) {
	RunWithCharacterGW(t, `
      types:
        - name: Query
          actions:
            - type: mount
              field: mysearch
              upstream: characters
              query:  "query($text: String!) { search(name:$text) }"`,
		func(gw, _ *httpgql.Client) {

			res := gw.ServeGraphQL(&graphql.Request{
				Query: `{ mysearch(text:"Rukia") { name { full } } }`,
			})
			require.NoError(t, res.Error())
			assert.Equal(t, `{"mysearch":{"name":{"full":"Rukia Kuchiki"}}}`, string(res.Data))

		})
}

func TestMountRootQueryOnNamedField(t *testing.T) {
	RunWithCharacterGW(t, `
      types:
        - name: Query
          actions:
            - type: mount
              field: charactersQuery
              upstream: characters
              query: query {}`,
		func(gw, char *httpgql.Client) {

			expected := map[string]interface{}{}
			err := char.Exec(ctx, &expected, `
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
			require.NoError(t, err)

			actual := map[string]interface{}{}
			err = gw.Exec(ctx, &actual, `
				query  {
				  charactersQuery {
					characters {
					  id
					  name {
						first
						last
						full
					  }
					}
				  }
				}`)

			require.NoError(t, err)
			assert.Equal(t, expected, actual["charactersQuery"])
		})
}

func TestMountAllFieldsOnRootQuery(t *testing.T) {
	RunWithCharacterGW(t, `
      types:
        - name: Query
          actions:
            - type: mount
              upstream: characters
              query: query {}`,
		func(gateway, characters *httpgql.Client) {

			expected := map[string]interface{}{}
			err := characters.Exec(ctx, &expected, `
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
			require.NoError(t, err)

			actual := map[string]interface{}{}
			err = gateway.Exec(ctx, &actual, `
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
			require.NoError(t, err)
			assert.Equal(t, expected, actual)
		})
}

func TestMountNamedFieldWithArguments(t *testing.T) {
	RunWithCharacterGW(t, `
      types:
        - name: Query
          actions:
            - type: mount
              field: mysearch
              upstream: characters
              query: query { search }`,
		func(gateway, characters *httpgql.Client) {
			res := gateway.ServeGraphQL(&graphql.Request{
				Query: `{ mysearch(name:"Rukia") { name { full } } }`,
			})
			require.NoError(t, res.Error())
			assert.Equal(t, `{"mysearch":{"name":{"full":"Rukia Kuchiki"}}}`, string(res.Data))
		})
}
