package gateway

import (
	"context"
	"encoding/json"
	"fmt"
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

	upstreamSelections, err := getSelectedFields(upstream.schema, upstreamQueryDoc, upstreamOp)
	if err != nil {
		return err
	}

	// find the result type of the upstream query.
	var upstreamResultType schema.Type = upstream.schema.EntryPoints[upstreamOp.Type]
	if upstreamResultType == nil {
		return errors.Errorf("The upstream does not have any %s entry points", upstreamOp.Type)
	}

	for _, s := range upstreamSelections {
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

	field.Args = []*schema.InputValue{}

	// query {} has no selections...
	if len(upstreamSelections) > 0 {
		lastSelection := upstreamSelections[len(upstreamSelections)-1]
		field.Type = lastSelection.Field.Type

		for _, arg := range lastSelection.Field.Args {
			if _, ok := lastSelection.Selection.Arguments.Get(arg.Name); ok {
				continue
			}
			field.Args = append(field.Args, arg)
		}
	}

	for _, value := range upstreamOp.Vars {
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

	c.Resolver.Set(c.Type.Name, field.Name, func(request *resolvers.ResolveRequest, _ resolvers.Resolution) resolvers.Resolution {

		loads := request.Context.Value(UpstreamLoadsContextKey).(UpstreamLoads)
		load := loads.loads[upstream.id]

		if load == nil {
			load = &UpstreamLoad{
				ctx:       request.Context,
				upstream:  upstream,
				variables: request.ExecutionContext.GetVars(),
			}
			if load.variables == nil {
				load.variables = map[string]interface{}{}
			}
			loads.loads[upstream.id] = load
		}

		requestDoc := request.ExecutionContext.GetDocument()
		requestOp := request.ExecutionContext.GetOperation()

		upstreamQueryDoc := upstreamQueryDoc.DeepCopy()
		upstreamOp := upstreamQueryDoc.Operations[0]

		// We need to join the upstream query args configured on the gateway with the client query args to make
		// a joined query.  Here is a kitchen sink example that can help you think of all the combination of ways
		// it they can be used.
		//
		// g1 = query upstream($a:A, $b:B, $c:C) { | query client($w:A, $x:B, $y:Y, $z:Z) { | query joined($w:A, $x:B, $y:Y, $z:Z) {
		//   f1(f1a:$a, f1b:$b, f1o:"a") {         |   g1(a:$w, b:$x, c:"c" y:$y)           |   f1(f1a:$w, f1b:$x f1o:"a") {
		//     f2(f2c:$c, f2o:"b")                 |     f3(z:$z)                           |     f2(f2c:"c", f2o:"b", y:$y) {
		//   }                                     |   }                                    |       f3(z:$z)
		// }                                       | }                                      | } } }

		// Example (1):
		// upstream:  mysearch => ($text: String!) { search(name:$text) }"
		// request :  { mysearch(text:"Rukia") { name { full } } }

		// find the leaf selection the upstream query...
		upstreamLeaf, upstreamLeafPath := getLeafAndResolveVars(upstreamQueryDoc, upstreamOp, requestDoc, request.Selection.Arguments)
		upstreamLeaf.SetSelections(upstreamQueryDoc, request.Selection.Selections)

		joinedOpVars := schema.InputValueList{}
		for _, v := range requestOp.Vars {
			c := *v
			c.Type = upstream.ToUpstreamType(c.Type)
			joinedOpVars = append(joinedOpVars, &c)
		}

		if upstreamLeaf, ok := upstreamLeaf.(*schema.FieldSelection); ok {
			extraArgs := map[string]schema.Argument{}
			for _, a := range request.Selection.Arguments {
				extraArgs[a.Name] = a
			}
			for _, arg := range upstreamOp.Vars {
				delete(extraArgs, strings.TrimPrefix(arg.Name, "$"))
			}
			for _, arg := range extraArgs {
				upstreamLeaf.Arguments = append(upstreamLeaf.Arguments, arg)
			}
		}

		upstreamOp.Vars = joinedOpVars
		upstreamQueryDoc.Fragments = requestDoc.Fragments
		load.selections = append(load.selections, upstreamQueryDoc)

		if upstreamOp.Type != schema.Subscription {
			return func() (reflect.Value, error) {

				if !loads.started {
					loads.started = true
					for _, load := range loads.loads {
						load.merged = mergeQueryDocs(load.selections)
						load.selections = nil
						// request.RunAsync handles limiting concurrency..
						request.RunAsync(load.resolution)()
					}
				}

				// we call this to make sure we wait for the async resolution to complete
				load.resolution()
				return getUpstreamValue(request.Context, load.response, load.merged, upstreamLeafPath)
			}
		} else {
			return func() (reflect.Value, error) {

				if !loads.started {
					loads.started = true
					for _, load := range loads.loads {
						load.merged = mergeQueryDocs(load.selections)
						load.selections = nil
					}
				}

				gqlRequest := &graphql.Request{
					Context:   load.ctx,
					Query:     load.merged.String(),
					Variables: load.variables,
				}

				stream := upstream.subscriptionClient(gqlRequest)
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

							v, err := getUpstreamValue(request.Context, result, upstreamQueryDoc, upstreamLeafPath)
							if err != nil {
								ctx.FireSubscriptionEvent(v, err)
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

func getLeafAndResolveVars(doc *schema.QueryDocument, from schema.Selection, requestDoc *schema.QueryDocument, args schema.ArgumentList) (schema.Selection, []schema.Selection) {
	path := []schema.Selection{}
	lastSelections := from.GetSelections(doc)
	for len(lastSelections) > 0 {
		from = lastSelections[0]
		path = append(path, from)
		lastSelections = from.GetSelections(doc)

		if field, ok := from.(*schema.FieldSelection); ok {
			for i, a := range field.Arguments {
				field.Arguments[i].Value = resolveVars(a.Value, args)
			}
		}
	}
	return from, path
}

func resolveVars(l schema.Literal, args schema.ArgumentList) schema.Literal {
	switch l := l.(type) {
	case *schema.Variable:
		if x, ok := args.Get(l.Name); ok {
			return x
		}
	case *schema.ObjectLit:
		for i, field := range l.Fields {
			l.Fields[i].Value = resolveVars(field.Value, args)
		}
	case *schema.ListLit:
		for i, entry := range l.Entries {
			l.Entries[i] = resolveVars(entry, args)
		}
	}
	return l
}

func (u *upstreamServer) ToUpstreamType(t schema.Type) schema.Type {
	switch t := t.(type) {
	case *schema.NonNull:
		return &schema.NonNull{OfType: u.ToUpstreamType(t.OfType)}
	case *schema.List:
		return &schema.List{OfType: u.ToUpstreamType(t.OfType)}
	case *schema.TypeName:
		name := t.Name
		name = u.gatewayToUpstreamTypeNames[name]
		return &schema.TypeName{
			Name: name,
		}
	case schema.NamedType:
		name := t.TypeName()
		name = u.gatewayToUpstreamTypeNames[name]
		return &schema.TypeName{
			Name: name,
		}
	}
	return t
}

func getUpstreamValue(ctx context.Context, result *graphql.Response, doc *schema.QueryDocument, selectionPath []schema.Selection) (reflect.Value, error) {
	if len(result.Errors) > 0 {
		return reflect.Value{}, result.Error()
	}

	data := map[string]interface{}{}
	err := json.Unmarshal(result.Data, &data)
	if err != nil {
		return reflect.Value{}, err
	}
	var v interface{} = data

	for _, sel := range selectionPath {
		switch sel := sel.(type) {
		case *schema.FieldSelection:
			if m, ok := v.(map[string]interface{}); ok {
				v = m[sel.Extension.(string)]
			} else {
				return reflect.Value{}, errors.Errorf("expected upstream field not found: %s", sel.Name)
			}
		}
	}

	// This enables the upstreamDomResolverInstance for all child fields of this result.
	// needed to property handle field aliases.
	return reflect.ValueOf(resolvers.ValueWithContext{
		Value:   reflect.ValueOf(v),
		Context: context.WithValue(ctx, upstreamDomResolverInstance, true),
	}), nil
	//return reflect.ValueOf(v), nil
}
