package starwars_characters

import (
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/chirino/graphql"
	"github.com/ghodss/yaml"
)

func New() *graphql.Engine {
	engine := graphql.New()
	err := engine.Schema.Parse(Schema)
	if err != nil {
		panic(err)
	}

	root := root{}
	err = yaml.Unmarshal([]byte(data), &root)
	if err != nil {
		panic(err)
	}

	for _, v := range root.Droids {
		v.self = v
	}
	for _, v := range root.Humans {
		v.self = v
	}

	engine.Root = root
	return engine
}

type root struct {
	Humans []*human `json:"humans"`
	Droids []*droid `json:"droids"`
}

// 		hero(episode: Episode = NEWHOPE): Character
func (r root) Hero(args struct{ Episode string }) *character {
	switch args.Episode {
	case "EMPIRE":
		return r.character("1000")
	default:
		return r.character("2001")
	}
}

//		search(text: String!): [SearchResult]!
func (r root) Search(args struct{ Text string }) (result []interface{}) {
	for _, h := range r.Humans {
		if strings.Contains(h.Name, args.Text) {
			result = append(result, h)
		}
	}
	for _, d := range r.Droids {
		if strings.Contains(d.Name, args.Text) {
			result = append(result, d)
		}
	}
	return
}

func (r root) characters(ids []string) (characters []character) {
	for _, id := range ids {
		if c := r.character(id); c != nil {
			characters = append(characters, *c)
		}
	}
	return
}

// character(id: ID!): Character
func (r root) Character(args struct{ Id string }) *character {
	return r.character(args.Id)
}

func (r root) character(id string) *character {
	h := r.human(id)
	if h != nil {
		return &h.character
	}
	d := r.droid(id)
	if d != nil {
		return &d.character
	}
	return nil
}

//		droid(id: ID!): Droid
func (r root) Droid(args struct{ Id string }) *droid {
	return r.droid(args.Id)
}
func (r root) droid(id string) *droid {
	for _, d := range r.Droids {
		if d.ID == id {
			return d
		}
	}
	return nil
}

//		human(id: ID!): Human
func (r root) Human(args struct{ Id string }) *human {
	return r.human(args.Id)
}

func (r root) human(id string) *human {
	for _, h := range r.Humans {
		if h.ID == id {
			return h
		}
	}
	return nil
}

func (r root) friendsConnection(ids []string, First *int32, After *string) (*friendsConnection, error) {
	from := 0
	if After != nil {
		b, err := base64.StdEncoding.DecodeString(string(*After))
		if err != nil {
			return nil, err
		}
		i, err := strconv.Atoi(strings.TrimPrefix(string(b), "cursor"))
		if err != nil {
			return nil, err
		}
		from = i
	}

	to := len(ids)
	if First != nil {
		to = from + int(*First)
		if to > len(ids) {
			to = len(ids)
		}
	}

	return &friendsConnection{
		ids:  ids,
		from: from,
		to:   to,
	}, nil
}

const ComplexStarWarsCharacterQuery = `
query HeroNameAndFriends($episode: Episode, $withoutFriends: Boolean!, $withFriends: Boolean!) {
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
    ... on Droid {
      primaryFunction
    }
    ... on Human {
      height
    }
  }
  search(text: "an") {
    __typename
    ... on Human {
      name
    }
    ... on Droid {
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

fragment comparisonFields on Character {
  name
  appearsIn
  friends {
    name
  }
}

fragment height on Human {
  height
}

fragment friendsFragment on Character {
  friends {
    name
  }
}
`
const ComplexStarWarsCharacterQueryResult = `{
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
        "name": "Mutation"
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
        "name": "Subscription"
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
        "name": "appearsIn",
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
        "name": "primaryFunction",
        "args": [],
        "type": {
          "name": "String",
          "kind": "SCALAR"
        }
      }
    ]
  }
}`
