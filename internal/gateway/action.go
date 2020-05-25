package gateway

import (
	"github.com/chirino/graphql"
	"github.com/chirino/graphql/resolvers"
	"github.com/chirino/graphql/schema"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type actionRunner struct {
	Gateway   *graphql.Engine
	Endpoints map[string]*upstreamServer
	Resolver  resolvers.TypeAndFieldResolver
	Type      *schema.Object
}

type ActionWrapper struct {
	Action interface{} `yaml:"-"`
}

func (h *ActionWrapper) UnmarshalYAML(unmarshal func(interface{}) error) error {
	discriminator := struct {
		Type string `yaml:"type"`
	}{}
	err := unmarshal(&discriminator)
	if err != nil {
		return err
	}

	var action interface{}
	switch discriminator.Type {
	case "mount":
		action = &Mount{}
	case "rename":
		action = &Rename{}
	case "remove":
		action = &Remove{}
	case "link":
		action = &Link{}
	default:
		return errors.New("invalid action type")
	}

	unmarshal(action)
	h.Action = action
	return nil
}

func (h ActionWrapper) MarshalYAML() (interface{}, error) {
	typeValue := ""
	if h.Action != nil {
		switch h.Action.(type) {
		case *Mount:
			typeValue = "mount"
		case *Rename:
			typeValue = "rename"
		case *Link:
			typeValue = "link"
		case *Remove:
			typeValue = "remove"
		}
	}

	marshal, err := yaml.Marshal(h.Action)
	if err != nil {
		return nil, err
	}

	values := yaml.MapSlice{}
	err = yaml.Unmarshal(marshal, &values)
	if err != nil {
		return nil, err
	}

	result := yaml.MapSlice{
		yaml.MapItem{Key: "type", Value: typeValue},
	}
	return append(result, values...), nil
}
