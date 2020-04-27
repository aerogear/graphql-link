package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql/exec"
	"github.com/chirino/graphql/qerrors"
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
	Name        string `json:"name"`
	Description string `json:"description"`
	Endpoint    string `json:"endpoint"`
	Query       string `json:"query"`
}

type Config struct {
	ConfigDirectory        string                  `json:"-"`
	DisableSchemaDownloads bool                    `json:"disable-schema-downloads"`
	EnabledSchemaStorage   bool                    `json:"enable-schema-storage"`
	Endpoints              map[string]EndpointInfo `json:"endpoints"`
	Schema                 map[string][]Field      `json:"schema"`
}

type endpoint struct {
	client func(request *graphql.EngineRequest) *graphql.EngineResponse
	schema *schema.Schema
	info   EndpointInfo
}

var validGraphQLIdentifierRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z_0-9]*$`)

func NewEngine(config Config) (*graphql.Engine, error) {

	if config.ConfigDirectory == "" {
		config.ConfigDirectory = "."
	}
	if config.EnabledSchemaStorage {
		os.MkdirAll(filepath.Join(config.ConfigDirectory, "endpoints"), 0755)
	}

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

	endpoints := map[string]*endpoint{}

	for eid, info := range config.Endpoints {
		c := relay.NewClient(info.URL)
		endpoints[eid] = &endpoint{
			info:   info,
			client: c.ServeGraphQL,
		}
	}

	for eid, endpoint := range endpoints {
		s, err := loadEndpointSchema(config, eid, endpoint)
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

	for typeName, fields := range config.Schema {
		object := root.Schema.Types[typeName]
		if object == nil {
			object = &schema.Object{Name: typeName}
		}
		if object, ok := object.(*schema.Object); ok {
			for _, fieldConfig := range fields {
				if endpoint, ok := endpoints[fieldConfig.Endpoint]; ok {

					var field *schema.Field
					if fieldConfig.Name != "" {
						field = &schema.Field{Name: fieldConfig.Name}
						if fieldConfig.Description != "" {
							field.Desc = &schema.Description{Text: fieldConfig.Description}
						}
					}

					err := Mount(root, object.Name, field, fieldResolver, endpoint.schema, endpoint.client, fieldConfig.Query)
					if err != nil {
						return nil, err
					}
				} else {
					return nil, errors.New("invalid endpoint id: " + fieldConfig.Endpoint)
				}
			}
		} else {
			return nil, errors.Errorf("can only configure fields on OBJECT types: %s is a %s", typeName, object.Kind())
		}
	}
	return root, nil
}

func loadEndpointSchema(config Config, eid string, endpoint *endpoint) (*schema.Schema, error) {

	schemaText := endpoint.info.Schema
	if strings.TrimSpace(schemaText) != "" {
		log.Printf("using static schema for endpoint %s: %s", eid, endpoint.info.URL)
		return Parse(schemaText)
	}

	endpointSchemaFile := filepath.Join(config.ConfigDirectory, "endpoints", eid+".graphql")
	endpointSchemaFileExists := false
	if stat, err := os.Stat(endpointSchemaFile); err == nil && !stat.IsDir() {
		endpointSchemaFileExists = true
	}

	if !config.DisableSchemaDownloads {
		log.Printf("downloading schema for endpoint %s: %s", eid, endpoint.info.URL)
		s, err := graphql.GetSchema(endpoint.client)

		if err != nil {
			if endpointSchemaFileExists {
				log.Printf("download failed (will load cached schema version): %v", err)
			} else {
				return nil, errors.Wrap(err, "download failed")
			}
		}

		// We may need to store it if it succeeded.
		if err == nil && config.EnabledSchemaStorage {
			err := ioutil.WriteFile(endpointSchemaFile, []byte(s.String()), 0644)
			if err != nil {
				return nil, errors.Wrap(err, "could not update schema")
			}
		}

		return s, nil
	}

	if endpointSchemaFileExists {
		log.Printf("loading previously stored schema: %s", endpointSchemaFile)
		// This could be a transient failure... see if we have previously save it's schema.
		data, err := ioutil.ReadFile(endpointSchemaFile)
		if err != nil {
			return nil, err
		}
		return Parse(string(data))
	}

	return nil, errors.Errorf("no schema defined for endpoint %s: %s", eid, endpoint.info.URL)
}

func Parse(schemaText string) (*schema.Schema, error) {
	s := schema.New()
	err := s.Parse(schemaText)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func Mount(root *graphql.Engine, rootTypeName string, rootField *schema.Field, resolver resolvers.TypeAndFieldResolver, childSchema *schema.Schema, serveGraphQL graphql.ServeGraphQLFunc, childQuery string) error {

	rootType := root.Schema.Types[rootTypeName].(*schema.Object)

	q, qerr := query.Parse(childQuery)
	if qerr != nil {
		return qerr
	}

	selections, err := GetSelectedFields(childSchema, q)
	if err != nil {
		return err
	}

	selectionAliases := []string{}
	for _, s := range selections {
		selectionAliases = append(selectionAliases, s.Selection.Alias.Text)
	}

	var querySplitter = regexp.MustCompile(`[}\s]*$`) //childSchema
	queryTail := querySplitter.FindString(childQuery)
	queryHead := strings.TrimSuffix(childQuery, queryTail)


	// We are mounting onto a single field...
	if rootField != nil {
		rootField.Type = rootType

		variablesUsed := map[string]*schema.InputValue{}
		for _, selection := range selections {
			for _, arg := range selection.Selection.Arguments {
				CollectVariablesUsed(variablesUsed, q.Operations[0], arg.Value)
			}
			rootField.Type = selection.Field.Type
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
		*field = *rootField

		resolver.Set(rootTypeName, rootField.Name, func(request *resolvers.ResolveRequest, _ resolvers.Resolution) resolvers.Resolution {
			return func() (reflect.Value, error) {

				clientQuery := &bytes.Buffer{}
				clientQuery.WriteString(queryHead)
				request.Selection.Selections.WriteTo(clientQuery)
				clientQuery.WriteString(queryTail)

				query := clientQuery.String()
				result := serveGraphQL(&graphql.Request{
					Context:   request.Context.GetContext(),
					Query:     query,
					Variables: request.Args,
				})

				return processResponse(result, query, selectionAliases)
			}
		})
	} else {
		// We are appending to the entire object
		operationType := q.Operations[0].Type
		childType := childSchema.EntryPoints[operationType].(*schema.Object)
		for _, s := range selections {
			if t, ok := s.Field.Type.(*schema.Object); ok {
				childType = t
			} else {
				return fmt.Errorf("a field name is reqired for the query selection")
			}
		}

		endpointResolver := func(request *resolvers.ResolveRequest, _ resolvers.Resolution) resolvers.Resolution {
			return func() (reflect.Value, error) {

				clientQuery := &bytes.Buffer{}
				clientQuery.WriteString(queryHead)
				request.Selection.WriteTo(clientQuery)
				clientQuery.WriteString(queryTail)

				query := clientQuery.String()
				result := serveGraphQL(&graphql.Request{
					Context:   request.Context.GetContext(),
					Query:     query,
					Variables: request.Args,
				})

				return processResponse(result, query, []string{request.Field.Name})
			}
		}

		for _, f := range childType.Fields {
			f.AddIfMissing(root.Schema, childSchema)
			if rootType.Fields.Get(f.Name) != nil {
				// Should we error out instead?
				continue
			}
			rootType.Fields = append(rootType.Fields, f)
			resolver.Set(rootType.Name, f.Name, endpointResolver)
		}

	}

	return nil
}

func processResponse(result *graphql.Response, query string, selectionAliases []string) (reflect.Value, error) {
	if len(result.Errors) > 0 {
		log.Println("query failed: ", query)
		return reflect.Value{}, result.Error()
	}

	data := map[string]interface{}{}
	err := json.Unmarshal(result.Data, &data)
	if err != nil {
		return reflect.Value{}, err
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

func CollectVariablesUsed(usedVariables map[string]*schema.InputValue, op *query.Operation, l schema.Literal) *graphql.Error {
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
			return qerrors.Errorf("variable name '%s' not found defined in operation arguments", l.Name).
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

	var result []exec.FieldSelection

	for len(selections) > 0 {
		fields, errs := fsc.Apply(selections)
		if len(errs) > 0 {
			return nil, errs.Error()
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
