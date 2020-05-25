package gateway

import (
	"github.com/chirino/graphql/schema"
	"github.com/pkg/errors"
)

type Link struct {
	Field       string            `yaml:"field,omitempty"`
	Description string            `yaml:"description,omitempty"`
	Upstream    string            `yaml:"upstream,omitempty"`
	Query       string            `yaml:"query,omitempty"`
	Vars        map[string]string `yaml:"vars,omitempty"`
}

func (c actionRunner) link(action *Link) error {
	endpoint, ok := c.Endpoints[action.Upstream]
	if !ok {
		return errors.New("invalid endpoint id: " + action.Upstream)
	}
	if action.Field == "" {
		return errors.New("field must be set")
	}
	field := schema.Field{Name: action.Field}
	if action.Description != "" {
		field.Desc = schema.Description{Text: action.Description}
	}
	return mount(c, field, endpoint, action.Query, action.Vars)
}
