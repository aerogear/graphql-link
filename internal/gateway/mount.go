package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql/qerrors"
	"github.com/chirino/graphql/resolvers"
	"github.com/chirino/graphql/schema"
	"github.com/pkg/errors"
)

type Mount struct {
	Action
	Field       string `json:"field"`
	Description string `json:"description"`
	Upstream    string `json:"upstream"`
	Query       string `json:"query"`
}

func (c actionRunner) mount(action *Mount) error {
	endpoint, ok := c.Endpoints[action.Upstream]
	if !ok {
		return errors.New("invalid endpoint id: " + action.Upstream)
	}
	field := schema.Field{Name: action.Field}
	if action.Description != "" {
		field.Desc = schema.Description{Text: action.Description}
	}
	err := mount(c, field, endpoint, action.Query)
	if err != nil {
		return err
	}
	return nil
}

var emptySelectionRegex = regexp.MustCompile(`{\s*}\s*$`)
var querySplitter = regexp.MustCompile(`[}\s]*$`)

func mount(c actionRunner, field schema.Field, upstream *upstreamServer, upstreamQuery string) error {

	upstreamQueryDoc := &schema.QueryDocument{}
	qerr := upstreamQueryDoc.Parse(upstreamQuery)
	if qerr != nil {
		return qerr
	}
	if len(upstreamQueryDoc.Operations) != 1 {
		return qerrors.New("query document can only contain one operation")
	}
	upstreamOp := upstreamQueryDoc.Operations[0]

	selections, err := getSelectedFields(upstream.schema, upstreamQueryDoc, upstreamOp)
	if err != nil {
		return err
	}

	// find the result type of the upstream query.
	var upstreamResultType schema.Type = upstream.schema.EntryPoints[upstreamOp.Type]
	if upstreamResultType == nil {
		return errors.Errorf("The upstream does not have any %s entry points", upstreamOp.Type)
	}

	for _, s := range selections {
		upstreamResultType = schema.DeepestType(s.Field.Type)
	}

	if field.Name == "" {

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
			err = mount(c, *f, upstream, upstreamQuery)
			if err != nil {
				return err
			}
		}
		return nil
	}
	field.Type = upstreamResultType

	variablesUsed := map[string]*schema.InputValue{}
	for _, selection := range selections {
		for _, arg := range selection.Selection.Arguments {
			err := collectVariablesUsed(variablesUsed, upstreamQueryDoc.Operations[0], arg.Value)
			if err != nil {
				return err
			}
		}
	}

	field.Args = []*schema.InputValue{}

	// query {} has no selections...
	if len(selections) > 0 {

		lastSelection := selections[len(selections)-1]
		field.Type = lastSelection.Field.Type

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
				field.Args = append(field.Args, arg)
			}
		}
	}

	for _, value := range variablesUsed {
		field.Args = append(field.Args, &schema.InputValue{
			Name: strings.TrimPrefix(value.Name, "$"),
			Type: value.Type,
		})
	}
	sort.Slice(field.Args, func(i, j int) bool {
		return field.Args[i].Name < field.Args[j].Name
	})

	// make sure the types of the upstream schema get added to the root schema
	field.AddIfMissing(c.Gateway.Schema, upstream.schema)
	for _, v := range field.Args {
		t, err := schema.ResolveType(v.Type, c.Gateway.Schema.Resolve)
		if err != nil {
			return err
		}
		v.Type = t
	}

	mountType := c.Gateway.Schema.Types[c.Type.Name].(*schema.Object)
	existingField := mountType.Fields.Get(field.Name)
	if existingField == nil {
		// create a field object if it does not exist...
		existingField = &schema.Field{}
		mountType.Fields = append(mountType.Fields, existingField)
	}
	// overwrite the field with the provided config
	*existingField = field

	selectionAliases := []string{}
	for _, s := range selections {
		selectionAliases = append(selectionAliases, s.Selection.Alias)
	}

	c.Resolver.Set(c.Type.Name, field.Name, func(request *resolvers.ResolveRequest, _ resolvers.Resolution) resolvers.Resolution {
		return func() (reflect.Value, error) {

			// reparse to avoid modifying the original.
			upstreamQueryDoc := &schema.QueryDocument{}
			upstreamQueryDoc.Parse(upstreamQuery)
			upstreamOp := upstreamQueryDoc.Operations[0]

			vars := schema.InputValueList{}
			for _, v := range upstreamOp.Vars {
				c := *v
				c.Type = upstream.ToUpstreamType(c.Type)
				vars = append(vars, &c)
			}

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
				c := schema.InputValue{
					Name: "$" + k,
					Type: upstream.ToUpstreamType(t),
				}
				vars = append(vars, &c)

				lastSelectionField := lastSelection.(*schema.FieldSelection)
				lastSelectionField.Arguments = append(lastSelectionField.Arguments, schema.Argument{
					Name:  k,
					Value: &schema.Variable{Name: k},
				})
			}
			upstreamOp.Vars = vars

			lastSelection.SetSelections(upstreamQueryDoc, request.Selection.Selections)
			query := upstreamQueryDoc.String()

			ggraphqlRequest := &graphql.Request{
				Context:   request.ExecutionContext.GetContext(),
				Query:     query,
				Variables: request.Args,
			}
			if upstreamOp.Type != schema.Subscription {

				result := upstream.client(ggraphqlRequest)

				v, err := getUpstreamValue(request.Context, result, selectionAliases)
				if err != nil {
					log.Println("query failed: ", query)
				}
				return v, err

			} else {
				stream := upstream.subscriptionClient(ggraphqlRequest)
				ctx := request.ExecutionContext
				go func() {
					for {
						select {
						case <-ctx.GetContext().Done():
							// This handles the case where the gateway client cancels the subscription...
							ctx.FireSubscriptionClose()
							return
						case result := <-stream:
							if result == nil {
								// the upstream closed before the client closed us
								ctx.FireSubscriptionClose()
								return
							}
							// We got data from the upstream...

							v, err := getUpstreamValue(request.Context, result, selectionAliases)
							if err != nil {
								ctx.FireSubscriptionClose()
								return
							}
							ctx.FireSubscriptionEvent(v, err)
						}
					}
				}()
				return reflect.Value{}, nil
			}

		}
	})
	return nil
}

func (u *upstreamServer) ToUpstreamType(t schema.Type) schema.Type {
	switch t := t.(type) {
	case *schema.NonNull:
		return &schema.NonNull{OfType: u.ToUpstreamType(t.OfType)}
	case *schema.List:
		return &schema.List{OfType: u.ToUpstreamType(t.OfType)}
	case schema.NamedType:
		name := t.TypeName()
		name = u.gatewayToUpstreamTypeNames[name]
		return &schema.TypeName{
			Name: name,
		}
	}
	return t
}

func getUpstreamValue(ctx context.Context, result *graphql.Response, selectionAliases []string) (reflect.Value, error) {
	if len(result.Errors) > 0 {
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
		Context: context.WithValue(ctx, upstreamDomResolverInstance, true),
	}), nil
}
