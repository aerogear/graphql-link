package graphql_test

import (
	"testing"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql-gw/internal/gateway"
	"github.com/chirino/graphql-gw/internal/gateway/examples/starwars_characters"
	"github.com/chirino/graphql/httpgql"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/require"
)

func BenchmarkParallelGatewayProxy(b *testing.B) {

	upstreamEngine := starwars_characters.New()
	upstreamServer, _ := gateway.StartServer("0.0.0.0", 0, upstreamEngine, gateway.NoLog)
	defer upstreamServer.Close()

	var config gateway.Config
	err := yaml.Unmarshal([]byte(`
      upstreams:
        starwars:
          prefix: SW_
          url: TODO
      types:
        - name: Query
          actions:
            - type: mount
              upstream: starwars
              query: query {}
`), &config)

	require.NoError(b, err)
	config.Upstreams["starwars"].Upstream.(*gateway.GraphQLUpstream).URL = upstreamServer.URL + "/graphql"
	gatewayEngine, err := gateway.New(config)
	require.NoError(b, err)

	gatewayServer, _ := gateway.StartServer("0.0.0.0", 0, gatewayEngine, gateway.NoLog)
	defer gatewayServer.Close()

	client := httpgql.NewClient(gatewayServer.URL + "/graphql")

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			//
			resp := client.ServeGraphQL(&graphql.Request{
				Query: query,
			})
			require.NoError(b, resp.Error())
		}
	})

}

var query = `
query {
  hero {
    id
    name
    friends {
      name
    }
  }
}`

var complexQuery = `
query HeroNameAndFriends($episode: SW_Episode, $withoutFriends: Boolean!, $withFriends: Boolean!) {
  hero {
    id
    name
    friends {
      name
    }
  }
  empireHerhero: hero(episode: EMPIRE) {
    name
  }
  jediHero: hero(episode: JEDI) {
    name
  }
  human(id: "1000") {
    name
    height(unit: FOOT)
  }
  leftComparison: hero(episode: EMPIRE) {
    ...comparisonFields
    ...height
  }
  rightComparison: hero(episode: JEDI) {
    ...comparisonFields
    ...height
  }
  heroNameAndFriends: hero(episode: $episode) {
    name
  }
  heroSkip: hero(episode: $episode) {
    name
    friends @skip(if: $withoutFriends) {
      name
    }
  }
  heroInclude: hero(episode: $episode) {
    name
    ...friendsFragment @include(if: $withFriends)
  }
  inlineFragments: hero(episode: $episode) {
    name
    ... on SW_Droid {
      primaryFunction
    }
    ... on SW_Human {
      height
    }
  }
  search(text: "an") {
    __typename
    ... on SW_Human {
      name
    }
    ... on SW_Droid {
      name
    }
  }
  heroConnections: hero {
    name
    friendsConnection {
      totalCount
      pageInfo {
        startCursor
        endCursor
        hasNextPage
      }
      edges {
        cursor
        node {
          name
        }
      }
    }
  }
  __schema {
    types {
      name
    }
  }
  __type(name: "Droid") {
    name
    fields {
      name
      args {
        name
        type {
          name
        }
        defaultValue
      }
      type {
        name
        kind
      }
    }
  }
}

fragment comparisonFields on SW_Character {
  name
  appearsIn
  friends {
    name
  }
}

fragment height on SW_Human {
  height
}

fragment friendsFragment on SW_Character {
  friends {
    name
  }
}
`
