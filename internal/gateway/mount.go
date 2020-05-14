package gateway

import (
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

	upstreamDoc := &schema.QueryDocument{}
	qerr := upstreamDoc.Parse(upstreamQuery)
	if qerr != nil {
		return qerr
	}
	if len(upstreamDoc.Operations) != 1 {
		return qerrors.New("query document can only contain one operation")
	}
	upstreamOp := upstreamDoc.Operations[0]

	upstreamSelections, err := getSelectedFields(upstream.schema, upstreamDoc, upstreamOp)
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

		dataLoaders := request.Context.Value(DataLoadersKey).(DataLoaders)
		dataLoader := dataLoaders.loaders[upstream.id]

		if dataLoader == nil {
			dataLoader = &UpstreamDataLoader{
				ctx:       request.Context,
				upstream:  upstream,
				variables: request.ExecutionContext.GetVars(),
			}
			if dataLoader.variables == nil {
				dataLoader.variables = map[string]interface{}{}
			}
			dataLoaders.loaders[upstream.id] = dataLoader
		}

		requestDoc := request.ExecutionContext.GetDocument()
		requestOp := request.ExecutionContext.GetOperation()

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

		// We will be building up the joined query off a copy the stream query
		joinedDoc := upstreamDoc.DeepCopy()
		joinedOp := joinedDoc.Operations[0]
		mountPoint, mountPointPath := getLeafAndResolveVars(joinedDoc, joinedOp, requestDoc, request.Selection.Arguments)
		request.Selection.Extension = mountPoint

		mountPoint.SetSelections(joinedDoc, request.Selection.Selections)
		addMountPointArgs(joinedOp, mountPoint, request)
		joinedOp.Vars = upstream.ToUpstreamInputValueList(requestOp.Vars)
		joinedDoc.Fragments = requestDoc.Fragments
		dataLoader.queryDocs = append(dataLoader.queryDocs, joinedDoc)

		if joinedOp.Type != schema.Subscription {
			return func() (reflect.Value, error) {

				if !dataLoaders.started {
					dataLoaders.started = true
					for _, load := range dataLoaders.loaders {
						if len(load.queryDocs) > 0 {

							load.mergedDoc = mergeQueryDocs(load.queryDocs) //.DeepCopy()
							operation := load.mergedDoc.Operations[0]
							operation.Selections = addTypeNames(operation.Selections)
							load.queryDocs = nil

							// request.RunAsync handles limiting concurrency..
							request.RunAsync(load.resolution)()
						}
					}
				}

				// we call this to make sure we wait for the async resolution to complete
				dataLoader.resolution()
				return getUpstreamValue(request.Context, dataLoader.response, dataLoader.mergedDoc, mountPointPath)
			}
		} else {
			return func() (reflect.Value, error) {

				if !dataLoaders.started {
					dataLoaders.started = true
					for _, load := range dataLoaders.loaders {
						load.mergedDoc = mergeQueryDocs(load.queryDocs)
						load.queryDocs = nil
					}
				}

				gqlRequest := &graphql.Request{
					Context:   dataLoader.ctx,
					Query:     dataLoader.mergedDoc.String(),
					Variables: dataLoader.variables,
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

							v, err := getUpstreamValue(request.Context, result, joinedDoc, mountPointPath)
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

func addTypeNames(from schema.SelectionList) schema.SelectionList {
	// if true {}
	needTypename := false
	haveTypename := false
	for _, s := range from {
		switch s := s.(type) {
		case *schema.FieldSelection:
			if s.Name == "__typename" {
				haveTypename = true
			}
			if len(s.Selections) == 0 {
				needTypename = true
			} else {
				s.Selections = addTypeNames(s.Selections)
			}
		case *schema.InlineFragment:
			needTypename = true
			s.Selections = addTypeNames(s.Selections)
		case *schema.FragmentSpread:
			needTypename = true
			//s.Selections = enrich(s.Selections)
		}
	}
	if needTypename && !haveTypename {
		from = append(from, &schema.FieldSelection{
			Alias: "t",
			Name:  "__typename",
		})
	}
	return from
}

func addMountPointArgs(joinedOp *schema.Operation, mountPoint schema.Selection, request *resolvers.ResolveRequest) {
	if mountPoint, ok := mountPoint.(*schema.FieldSelection); ok {
		extraArgs := map[string]schema.Argument{}
		for _, a := range request.Selection.Arguments {
			extraArgs[a.Name] = a
		}
		for _, arg := range joinedOp.Vars {
			delete(extraArgs, strings.TrimPrefix(arg.Name, "$"))
		}
		for _, arg := range extraArgs {
			mountPoint.Arguments = append(mountPoint.Arguments, arg)
		}
	}
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
