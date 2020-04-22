package gateway

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"
	"strings"

	"github.com/chirino/graphql"
	qerrors "github.com/chirino/graphql/errors"
	"github.com/chirino/graphql/exec"
	"github.com/chirino/graphql/query"
	"github.com/chirino/graphql/relay"
	"github.com/chirino/graphql/resolvers"
	"github.com/chirino/graphql/schema"
	"github.com/pkg/errors"
)

type EndpointInfo struct {
	URL    string `json:"url"`
	Prefix string `json:"prefix"`
	Suffix string `json:"suffix"`
	Schema string `json:"schema"`
}

type Field struct {
	Description string `json:"description"`
	Endpoint    string `json:"endpoint"`
	Query       string `json:"query"`
}

type Config struct {
	Endpoints map[string]EndpointInfo `json:"endpoints"`
	Query     map[string]Field        `json:"query"`
	Mutation  map[string]Field        `json:"mutation"`
}

func NewEngine(config Config) (*graphql.Engine, error) {
	fieldResolver := resolvers.TypeAndFieldResolver{}
	root := graphql.New()
	err := root.Schema.Parse(`
schema {
    query: Query
    mutation: Mutation
}
type Query {}
type Mutation {}
`)
	if err != nil {
		panic(err)
	}
	root.Resolver = resolvers.List(root.Resolver, fieldResolver)

	type Endpoint struct {
		client func(request *graphql.EngineRequest) *graphql.EngineResponse
		schema *schema.Schema
		info   EndpointInfo
	}

	endpoints := map[string]*Endpoint{}

	for eid, info := range config.Endpoints {
		c := relay.NewClient(info.URL)
		endpoints[eid] = &Endpoint{
			info:   info,
			client: c.Post,
		}
	}

	for eid, endpoint := range endpoints {

		schemaText := endpoint.info.Schema
		if schemaText == "" {
			schemaText, err = GetSchema(endpoint.client)
			if err != nil {
				return nil, err
			}
		}

		s := schema.New()
		err = s.Parse(schemaText)
		if err != nil {
			return nil, err
		}

		if endpoint.info.Prefix != "" {
			s.RenameTypes(func(x string) string { return endpoint.info.Prefix + x })
		}
		if endpoint.info.Suffix != "" {
			s.RenameTypes(func(x string) string { return x + endpoint.info.Suffix })
		}

		endpoints[eid].schema = s
	}

	object := root.Schema.Types["Query"].(*schema.Object)
	for fieldName, fieldConfig := range config.Query {

		field := &schema.Field{}
		if fieldConfig.Description != "" {
			field.Desc = &schema.Description{Text: fieldConfig.Description}
		}
		field.Name = fieldName

		if endpoint, ok := endpoints[fieldConfig.Endpoint]; ok {
			err := Mount(root, object.Name, *field, fieldResolver, endpoint.schema, endpoint.client, fieldConfig.Query)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("invalid endpoint id: " + fieldConfig.Endpoint)
		}
	}
	return root, nil
}

func GetSchema(api graphql.StandardAPI) (string, error) {
	r := api(&graphql.EngineRequest{
		Context: context.Background(),
		Query:   "query{schema}",
	})
	if r.Error() != nil {
		return "", r.Error()
	}

	data := struct {
		Schema string `json:"schema"`
	}{}
	err := json.Unmarshal(r.Data, &data)
	if err != nil {
		return "", err
	}
	return data.Schema, nil
}

func Mount(root *graphql.Engine, rootTypeName string, rootField schema.Field, resolver resolvers.TypeAndFieldResolver, childSchema *schema.Schema, client graphql.StandardAPI, childQuery string) error {

	rootType := root.Schema.Types[rootTypeName].(*schema.Object)

	q, qerr := query.Parse(childQuery)
	if qerr != nil {
		return qerr
	}

	selections, err := GetSelectedFields(childSchema, q)
	if err != nil {
		return err
	}

	lastSelection := selections[len(selections)-1]
	rootField.Type = lastSelection.Field.Type

	variablesUsed := map[string]*schema.InputValue{}
	for _, selection := range selections {
		for _, arg := range selection.Selection.Arguments {
			CollectVariablesUsed(variablesUsed, q.Operations[0], arg.Value)
		}
	}

	rootField.Args = schema.InputValueList{}
	for _, value := range variablesUsed {
		rootField.Args = append(rootField.Args, value)
	}
	sort.Slice(rootField.Args, func(i, j int) bool {
		return rootField.Args[i].Name.Text < rootField.Args[j].Name.Text
	})

	// make sure the types of the child schema get added to the root schema
	rootField.AddIfMissing(root.Schema, childSchema)
	for _, v := range rootField.Args {
		t, err := schema.ResolveType(v.Type, root.Schema.Resolve)
		if err != nil {
			return err
		}
		v.Type = t
	}

	field := rootType.Fields.Get(rootField.Name)
	if field == nil {
		// create a field object if it does not exist...
		field = &schema.Field{}
		rootType.Fields = append(rootType.Fields, field)
	}
	// overwrite the field with the provided config
	*field = rootField

	selectionAliases := []string{}
	for _, s := range selections {
		selectionAliases = append(selectionAliases, s.Selection.Alias.Text)
	}

	resolver.Set(rootTypeName, rootField.Name, func(request *resolvers.ResolveRequest, _ resolvers.Resolution) resolvers.Resolution {
		return func() (reflect.Value, error) {

			result := client(&graphql.EngineRequest{
				Context:   request.Context.GetContext(),
				Query:     childQuery,
				Variables: request.Args,
			})

			data := map[string]interface{}{}
			if result.Data == nil {
				err := result.Error()
				if err == nil {
					err = errors.New("no json result provided")
				}
				return reflect.Value{}, err
			} else {
				err := json.Unmarshal(result.Data, &data)
				if err != nil {
					return reflect.Value{}, err
				}
			}

			var r interface{} = data
			for _, alias := range selectionAliases {
				if m, ok := r.(map[string]interface{}); ok {
					r = m[alias]
				} else {
					return reflect.Value{}, errors.Errorf("expected json field not found: " + strings.Join(selectionAliases, "."))
				}
			}
			return reflect.ValueOf(r), result.Error()
		}
	})
	return nil
}

func CollectVariablesUsed(usedVariables map[string]*schema.InputValue, op *query.Operation, l schema.Literal) *qerrors.QueryError {
	switch l := l.(type) {
	case *schema.ObjectLit:
		for _, f := range l.Fields {
			err := CollectVariablesUsed(usedVariables, op, f.Value)
			if err != nil {
				return err
			}
		}
	case *schema.ListLit:
		for _, entry := range l.Entries {
			err := CollectVariablesUsed(usedVariables, op, entry)
			if err != nil {
				return err
			}
		}
	case *schema.Variable:
		v := op.Vars.Get(l.Name)
		if v == nil {
			return qerrors.
				Errorf("variable name '%s' not found defined in operation arguments", l.Name).
				WithLocations(l.Loc).
				WithStack()
		}
		usedVariables[l.Name] = v
	}
	return nil
}

func GetSelectedFields(childSchema *schema.Schema, q *query.Document) ([]exec.FieldSelection, error) {
	if len(q.Operations) != 1 {
		return nil, qerrors.New("query document can only contain one operation")
	}
	op := q.Operations[0]
	onType := childSchema.EntryPoints[op.Type]

	fsc := exec.FieldSelectionContext{
		Path:          []string{},
		Schema:        childSchema,
		QueryDocument: q,
		OnType:        onType,
	}
	selections := op.Selections
	if len(selections) == 0 {
		return nil, qerrors.New("No selections").WithStack()
	}

	var result []exec.FieldSelection

	for len(selections) > 0 {
		fields, errs := fsc.Apply(selections)
		if len(errs) > 0 {
			return nil, qerrors.AsMulti(errs)
		}

		firstSelection := selections[0]
		if len(fields) == 0 {
			return nil, qerrors.New("No fields selected").WithLocations(firstSelection.Location()).WithStack()
		}
		if len(fields) > 1 {
			return nil, qerrors.New("please only select one field").WithLocations(firstSelection.Location()).WithStack()
		}
		result = append(result, fields[0])

		fsc.Path = append(fsc.Path, fields[0].Selection.Alias.Text)
		fsc.OnType = fields[0].Field.Type
		selections = fields[0].Selection.Selections
	}
	return result, nil
}
