package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql/qerrors"
	"github.com/chirino/graphql/resolvers"
	"github.com/chirino/graphql/schema"
	"github.com/pkg/errors"
)

func mount(gateway *graphql.Engine, mountTypeName string, mountField schema.Field, resolver resolvers.TypeAndFieldResolver, upstreamSchema *schema.Schema, serveGraphQL graphql.ServeGraphQLFunc, upstreamQuery string) error {

	upstreamQueryDoc := &schema.QueryDocument{}
	qerr := upstreamQueryDoc.Parse(upstreamQuery)
	if qerr != nil {
		return qerr
	}
	if len(upstreamQueryDoc.Operations) != 1 {
		return qerrors.New("query document can only contain one operation")
	}
	upstreamOp := upstreamQueryDoc.Operations[0]

	selections, err := getSelectedFields(upstreamSchema, upstreamQueryDoc, upstreamOp)
	if err != nil {
		return err
	}

	// find the result type of the upstream query.
	var upstreamResultType schema.Type = upstreamSchema.EntryPoints[upstreamOp.Type]
	for _, s := range selections {
		upstreamResultType = schema.DeepestType(s.Field.Type)
	}

	if mountField.Name == "" {

		fields := schema.FieldList{}

		// Get all the field names from it and mount them...
		switch upstreamResultType := upstreamResultType.(type) {
		case *schema.Object:
			fields = upstreamResultType.Fields
		case *schema.Interface:
			fields = upstreamResultType.Fields
		default:
			return errors.Errorf("Type '%s' does not have any fields to mount", upstreamResultType.String())
		}

		queryTail := ""
		queryHead := ""
		if emptySelectionRegex.MatchString(upstreamQuery) {
			queryHead = emptySelectionRegex.ReplaceAllString(upstreamQuery, "")
		} else {
			queryTail = querySplitter.FindString(upstreamQuery)
			queryHead = strings.TrimSuffix(upstreamQuery, queryTail)
		}

		for _, f := range fields {
			upstreamQuery = fmt.Sprintf("%s { %s } %s", queryHead, f.Name, queryTail)
			err = mount(gateway, mountTypeName, *f, resolver, upstreamSchema, serveGraphQL, upstreamQuery)
			if err != nil {
				return err
			}
		}
		return nil
	}
	mountField.Type = upstreamResultType

	variablesUsed := map[string]*schema.InputValue{}
	for _, selection := range selections {
		for _, arg := range selection.Selection.Arguments {
			err := collectVariablesUsed(variablesUsed, upstreamQueryDoc.Operations[0], arg.Value)
			if err != nil {
				return err
			}
		}
	}

	mountField.Args = []*schema.InputValue{}

	// query {} has no selections...
	if len(selections) > 0 {

		lastSelection := selections[len(selections)-1]
		mountField.Type = lastSelection.Field.Type

		for _, arg := range lastSelection.Field.Args {
			if variablesUsed[arg.Name] != nil {
				continue
			}
			for _, arg := range lastSelection.Field.Args {
				if lit, ok := lastSelection.Selection.Arguments.Get(arg.Name); ok {
					v := map[string]*schema.InputValue{}
					collectVariablesUsed(v, upstreamQueryDoc.Operations[0], lit)
					if len(v) != 0 {
						continue
					}
				}
				mountField.Args = append(mountField.Args, arg)
			}
		}
	}

	for _, value := range variablesUsed {
		mountField.Args = append(mountField.Args, &schema.InputValue{
			Name: strings.TrimPrefix(value.Name, "$"),
			Type: value.Type,
		})
	}
	sort.Slice(mountField.Args, func(i, j int) bool {
		return mountField.Args[i].Name < mountField.Args[j].Name
	})

	// make sure the types of the upstream schema get added to the root schema
	mountField.AddIfMissing(gateway.Schema, upstreamSchema)
	for _, v := range mountField.Args {
		t, err := schema.ResolveType(v.Type, gateway.Schema.Resolve)
		if err != nil {
			return err
		}
		v.Type = t
	}

	mountType := gateway.Schema.Types[mountTypeName].(*schema.Object)
	field := mountType.Fields.Get(mountField.Name)
	if field == nil {
		// create a field object if it does not exist...
		field = &schema.Field{}
		mountType.Fields = append(mountType.Fields, field)
	}
	// overwrite the field with the provided config
	*field = mountField

	selectionAliases := []string{}
	for _, s := range selections {
		selectionAliases = append(selectionAliases, s.Selection.Alias)
	}

	resolver.Set(mountTypeName, mountField.Name, func(request *resolvers.ResolveRequest, _ resolvers.Resolution) resolvers.Resolution {
		return func() (reflect.Value, error) {

			// reparse to avoid modifying the original.
			upstreamQueryDoc := &schema.QueryDocument{}
			upstreamQueryDoc.Parse(upstreamQuery)
			upstreamOp := upstreamQueryDoc.Operations[0]

			// find the leaf selection the upstream query...
			lastSelection := schema.Selection(upstreamOp)
			lastSelections := lastSelection.GetSelections(upstreamQueryDoc)
			for len(lastSelections) > 0 {
				lastSelection = lastSelections[0]
				lastSelections = lastSelection.GetSelections(upstreamQueryDoc)
			}
			// lets figure out what variables we need to add to the query...
			argsToAdd := map[string]schema.Type{}
			for _, arg := range request.Field.Args {
				argsToAdd[arg.Name] = arg.Type
			}
			for _, arg := range upstreamOp.Vars {
				delete(argsToAdd, strings.TrimPrefix(arg.Name, "$"))
			}

			for k, t := range argsToAdd {
				upstreamOp.Vars = append(upstreamOp.Vars, &schema.InputValue{
					Name: "$" + k,
					Type: t,
				})
				lastSelectionField := lastSelection.(*schema.FieldSelection)
				lastSelectionField.Arguments = append(lastSelectionField.Arguments, schema.Argument{
					Name:  k,
					Value: &schema.Variable{Name: k},
				})
			}

			lastSelection.SetSelections(upstreamQueryDoc, request.Selection.Selections)
			query := upstreamQueryDoc.String()

			result := serveGraphQL(&graphql.Request{
				Context:   request.ExecutionContext.GetContext(),
				Query:     query,
				Variables: request.Args,
			})

			if len(result.Errors) > 0 {
				log.Println("query failed: ", query)
				return reflect.Value{}, result.Error()
			}

			data := map[string]interface{}{}
			err := json.Unmarshal(result.Data, &data)
			if err != nil {
				return reflect.Value{}, err
			}

			var v interface{} = data
			for _, alias := range selectionAliases {
				if m, ok := v.(map[string]interface{}); ok {
					v = m[alias]
				} else {
					return reflect.Value{}, errors.Errorf("expected json field not found: " + strings.Join(selectionAliases, "."))
				}
			}

			// This enables the upstreamDomResolverInstance for all child fields of this result.
			// needed to property handle field aliases.
			return reflect.ValueOf(resolvers.ValueWithContext{
				Value:   reflect.ValueOf(v),
				Context: context.WithValue(request.Context, upstreamDomResolverInstance, true),
			}), nil
		}
	})
	return nil
}
