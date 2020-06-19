package gateway

import (
	"github.com/chirino/graphql-4-apis/pkg/apis"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type UpstreamWrapper struct {
	Upstream interface{} `yaml:"-"`
}

func (h *UpstreamWrapper) UnmarshalYAML(unmarshal func(interface{}) error) error {
	discriminator := struct {
		Type string `yaml:"type"`
	}{}
	err := unmarshal(&discriminator)
	if err != nil {
		return err
	}

	var upstream interface{}
	switch discriminator.Type {
	case "":
		upstream = &GraphQLUpstream{}
	case "graphql":
		upstream = &GraphQLUpstream{}
	case "openapi":
		upstream = &OpenApiUpstream{}
	default:
		return errors.New("invalid action type")
	}

	unmarshal(upstream)
	h.Upstream = upstream
	return nil
}

func (h UpstreamWrapper) MarshalYAML() (interface{}, error) {
	typeValue := ""
	if h.Upstream != nil {
		switch h.Upstream.(type) {
		case *GraphQLUpstream:
			typeValue = "graphql"
		case *OpenApiUpstream:
			typeValue = "openapi"
		}
	}

	marshal, err := yaml.Marshal(h.Upstream)
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

type UpstreamInfo struct {
	URL    string `yaml:"url,omitempty"`
	Prefix string `yaml:"prefix,omitempty"`
	Suffix string `yaml:"suffix,omitempty"`
	Schema string `yaml:"types,omitempty"`
}

type GraphQLUpstream struct {
	URL    string `yaml:"url,omitempty"`
	Prefix string `yaml:"prefix,omitempty"`
	Suffix string `yaml:"suffix,omitempty"`
	Schema string `yaml:"types,omitempty"`
}

type OpenApiUpstream struct {
	Openapi apis.EndpointOptions `yaml:"spec,omitempty"`
	APIBase apis.EndpointOptions `yaml:"api,omitempty"`
	Prefix  string               `yaml:"prefix,omitempty"`
	Suffix  string               `yaml:"suffix,omitempty"`
}
