package gateway

import (
	"github.com/chirino/graphql/schema"
	"github.com/pkg/errors"
)

type Remove struct {
	Field string `yaml:"field,omitempty"`
}

func (c actionRunner) remove(action *Remove) error {
	if action.Field == "" {
		return errors.New("field must be set")
	}
	c.Type.Fields = c.Type.Fields.Select(func(d *schema.Field) bool {
		return d.Name != action.Field
	})
	return nil
}
