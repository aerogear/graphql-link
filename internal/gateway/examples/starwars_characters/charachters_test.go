package starwars_characters_test

import (
	"encoding/json"
	"testing"

	"github.com/chirino/graphql-gw/internal/gateway/examples/starwars_characters"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {

	engine := starwars_characters.New()
	result := json.RawMessage{}
	engine.Exec(nil, &result, starwars_characters.ComplexStarWarsCharacterQuery, "episode", "JEDI",
		"withoutFriends", true,
		"withFriends", false)

	assert.JSONEq(t, `
{
  "hero": {
    "id": "2001",
    "name": "R2-D2",
    "friends": [
      {
        "name": "Luke Skywalker"
      },
      {
        "name": "Han Solo"
      },
      {
        "name": "Leia Organa"
      }
    ]
  },
  "empireHerhero": {
    "name": "Luke Skywalker"
  },
  "jediHero": {
    "name": "R2-D2"
  },
  "human": {
    "name": "Luke Skywalker",
    "height": 1.72
  },
  "leftComparison": {
    "name": "Luke Skywalker",
    "appearsIn": [
      "NEWHOPE",
      "EMPIRE",
      "JEDI"
    ],
    "friends": [
      {
        "name": "Han Solo"
      },
      {
        "name": "Leia Organa"
      },
      {
        "name": "C-3PO"
      },
      {
        "name": "R2-D2"
      }
    ],
    "height": 1.72
  },
  "rightComparison": {
    "name": "R2-D2",
    "appearsIn": [
      "NEWHOPE",
      "EMPIRE",
      "JEDI"
    ],
    "friends": [
      {
        "name": "Luke Skywalker"
      },
      {
        "name": "Han Solo"
      },
      {
        "name": "Leia Organa"
      }
    ]
  },
  "heroNameAndFriends": {
    "name": "R2-D2"
  },
  "heroSkip": {
    "name": "R2-D2"
  },
  "heroInclude": {
    "name": "R2-D2"
  },
  "inlineFragments": {
    "name": "R2-D2",
    "primaryFunction": "Astromech"
  },
  "search": [
    {
      "__typename": "Human",
      "name": "Han Solo"
    },
    {
      "__typename": "Human",
      "name": "Leia Organa"
    }
  ],
  "heroConnections": {
    "name": "R2-D2",
    "friendsConnection": {
      "totalCount": 3,
      "pageInfo": {
        "startCursor": "Y3Vyc29yMQ==",
        "endCursor": "Y3Vyc29yMw==",
        "hasNextPage": false
      },
      "edges": [
        {
          "cursor": "Y3Vyc29yMQ==",
          "node": {
            "name": "Luke Skywalker"
          }
        },
        {
          "cursor": "Y3Vyc29yMg==",
          "node": {
            "name": "Han Solo"
          }
        },
        {
          "cursor": "Y3Vyc29yMw==",
          "node": {
            "name": "Leia Organa"
          }
        }
      ]
    }
  },
  "__schema": {
    "types": [
      {
        "name": "Boolean"
      },
      {
        "name": "Character"
      },
      {
        "name": "Droid"
      },
      {
        "name": "Episode"
      },
      {
        "name": "Float"
      },
      {
        "name": "FriendsConnection"
      },
      {
        "name": "FriendsEdge"
      },
      {
        "name": "Human"
      },
      {
        "name": "ID"
      },
      {
        "name": "Int"
      },
      {
        "name": "LengthUnit"
      },
      {
        "name": "PageInfo"
      },
      {
        "name": "Query"
      },
      {
        "name": "SearchResult"
      },
      {
        "name": "Starship"
      },
      {
        "name": "String"
      },
      {
        "name": "__Directive"
      },
      {
        "name": "__DirectiveLocation"
      },
      {
        "name": "__EnumValue"
      },
      {
        "name": "__Field"
      },
      {
        "name": "__InputValue"
      },
      {
        "name": "__Schema"
      },
      {
        "name": "__Type"
      },
      {
        "name": "__TypeKind"
      }
    ]
  },
  "__type": {
    "name": "Droid",
    "fields": [
      {
        "name": "id",
        "args": [],
        "type": {
          "name": null,
          "kind": "NON_NULL"
        }
      },
      {
        "name": "name",
        "args": [],
        "type": {
          "name": null,
          "kind": "NON_NULL"
        }
      },
      {
        "name": "friends",
        "args": [],
        "type": {
          "name": null,
          "kind": "LIST"
        }
      },
      {
        "name": "friendsConnection",
        "args": [
          {
            "name": "first",
            "type": {
              "name": "Int"
            },
            "defaultValue": null
          },
          {
            "name": "after",
            "type": {
              "name": "ID"
            },
            "defaultValue": null
          }
        ],
        "type": {
          "name": null,
          "kind": "NON_NULL"
        }
      },
      {
        "name": "appearsIn",
        "args": [],
        "type": {
          "name": null,
          "kind": "NON_NULL"
        }
      },
      {
        "name": "primaryFunction",
        "args": [],
        "type": {
          "name": "String",
          "kind": "SCALAR"
        }
      }
    ]
  }
}`, string(result))
}
