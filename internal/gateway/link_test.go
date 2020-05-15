package gateway_test

import (
	"encoding/json"
	"testing"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql/httpgql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLink(t *testing.T) {
	RunWithCharacterGW(t, `
      types:
        - name: Query
          actions:
            - type: mount
              upstream: characters
              query: query {}
        - name: Character_t1
          actions:
            - type: link
              upstream: characters
              field: bf
              vars:
                $id: bestFriend
              query: |
                query ($id:String) { search(name:$id) }
`,
		func(gateway, characters *httpgql.Client) {
			res := gateway.ServeGraphQL(&graphql.Request{
				Variables: json.RawMessage(`{"episode":"JEDI", "withoutFriends": true, "withFriends": false}`),
				Query: `
{
	rukia: search(name:"Rukia") {
		bf { name{ full } }
	}
}
`,
			})
			require.NoError(t, res.Error())
			assert.Equal(t, `{"rukia":{"bf":{"name":{"full":"Ichigo Kurosaki"}}}}`, string(res.Data))
		})

}
