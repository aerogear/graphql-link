package gateway

import (
	"github.com/chirino/graphql/schema"
	"github.com/pkg/errors"
)

type Link struct {
	Action
	Field       string            `json:"field"`
	Description string            `json:"description"`
	Upstream    string            `json:"upstream"`
	Query       string            `json:"query"`
	Vars        map[string]string `json:"vars"`
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
