package gateway

import (
	"github.com/chirino/graphql"
	"github.com/chirino/graphql/qerrors"
	"github.com/chirino/graphql/schema"
)

func collectVariablesUsed(usedVariables map[string]*schema.InputValue, op *schema.Operation, l schema.Literal) *graphql.Error {
	switch l := l.(type) {
	case *schema.ObjectLit:
		for _, f := range l.Fields {
			err := collectVariablesUsed(usedVariables, op, f.Value)
			if err != nil {
				return err
			}
		}
	case *schema.ListLit:
		for _, entry := range l.Entries {
			err := collectVariablesUsed(usedVariables, op, entry)
			if err != nil {
				return err
			}
		}
	case *schema.Variable:
		v := op.Vars.Get(l.String())
		if v == nil {
			return qerrors.Errorf("variable name '%s' not found defined in operation arguments", l.Name).
				WithLocations(l.Loc).
				WithStack()
		}
		usedVariables[l.Name] = v
	}
	return nil
}
