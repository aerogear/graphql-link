package shows

import (
	"strings"

	"github.com/chirino/graphql"
	"github.com/ghodss/yaml"
)

const data = `
shows:
- id: 1
  name: Bleach
  characters:
  - Rukia Kuchiki
  - Ichigo Kurosaki
  - Orihime Inoue
- id: 2
  name: Sword Art Online
  characters:
  - Kirito
  - Eugeo
  - Alice Schuberg
`

var Schema = `
	schema {
		query: Query
	}
	type Query {
		shows: [Show!]!
		search(name:String!): Show
	}
	type Show {
		id: ID!
		name: String
		characters: [String!]!
	}
`

type root struct {
	Shows []shows `json:"shows"`
}

func (r root) Search(args struct{ Name string }) *shows {
	name := strings.ToLower(args.Name)
	for _, x := range r.Shows {
		if strings.Contains(strings.ToLower(x.Name), name) {
			return &x
		}
		for _, y := range x.Characters {
			if strings.Contains(strings.ToLower(y), name) {
				return &x
			}
		}
	}
	return nil
}

type shows struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Characters []string `json:"characters"`
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
