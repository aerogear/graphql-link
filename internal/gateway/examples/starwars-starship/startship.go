package starwars_starship

import (
	"strings"

	"github.com/chirino/graphql"
	"github.com/ghodss/yaml"
)

const data = `
characters:
- id: 1
  name:
    first: Rukia
    last: Kuchiki
- id: 2
  name:
    first: Ichigo
    last: Kurosaki
- id: 3
  name:
    first: Orihime
    last: Inoue
- id: 3
  name:
   first: Kirito
- id: 3
  name:
    first: Eugeo
- id: 3
  name:
    first: Alice
    last: Schuberg
`

var Schema = `
	schema {
		query: Query
	}
	type Query {
		characters: [Character!]!
		search(name:String!): Character
	}
	type Character {
		id: ID!
		name: Name
	}
	type Name {
		first: String
		last: String
		full: String
	}
`

type root struct {
	Characters []character `json:"characters"`
}

func (r root) Search(args struct{ Name string }) *character {
	name := strings.ToLower(args.Name)
	for _, x := range r.Characters {
		if strings.Contains(strings.ToLower(x.Name.Full()), name) {
			return &x
		}
	}
	return nil
}

type character struct {
	ID   string `json:"id"`
	Name name   `json:"name"`
}

type name struct {
	First string `json:"first"`
	Last  string `json:"last"`
}

func (r name) Full() string {
	return r.First + " " + r.Last
}

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
	engine.Root = root
	return engine
}
