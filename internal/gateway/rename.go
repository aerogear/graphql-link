package gateway

import (
	"github.com/chirino/graphql/resolvers"
	"github.com/pkg/errors"
)

type Rename struct {
	Field string `yaml:"field,omitempty"`
	To    string `yaml:"to,omitempty"`
}

func (c actionRunner) rename(action *Rename) error {
	to := action.To
	if to == "" {
		return errors.New("rename action to field not set.")
	}
	if action.Field == "" {
		// we are rename the type...

		fromType := c.Type.Name
		namedType := c.Gateway.Schema.Types[to]
		if namedType != nil {
			return errors.Errorf("cannot rename type '%s'; to '%s': it already exists", fromType, to)
		}

		c.Type.Name = to
		c.Gateway.Schema.Types[to] = c.Type

		delete(c.Gateway.Schema.Types, fromType)

		// rename the resolvers...
		for key, fn := range c.Resolver {
			if key.Type == fromType {
				delete(c.Resolver, key)
				key.Type = to
				c.Resolver[key] = fn
			}
		}

	} else {
		// we are renaming a field.
		fromType := c.Type.Name
		fromField := action.Field

		field := c.Type.Fields.Get(fromField)
		if field == nil {
			return errors.Errorf("cannot rename field '%s' to '%s': from field does not exist", fromField, to)
		}

		if c.Type.Fields.Get(to) != nil {
			return errors.Errorf("cannot rename field '%s' to '%s': to field already exists", fromField, to)
		}
		field.Name = to

		// rename the resolver, if there is one...
		key := resolvers.TypeAndFieldKey{
			Type:  fromType,
			Field: fromField,
		}
		fn := c.Resolver[key]
		if fn != nil {
			delete(c.Resolver, key)
			key.Field = to
			c.Resolver[key] = fn
		}
	}
	return nil
}
