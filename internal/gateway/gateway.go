package gateway

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql/exec"
	"github.com/chirino/graphql/qerrors"
	"github.com/chirino/graphql/relay"
	"github.com/chirino/graphql/resolvers"
	"github.com/chirino/graphql/schema"
	"github.com/pkg/errors"
)

type EndpointInfo struct {
	URL    string `json:"url"`
	Prefix string `json:"prefix"`
	Suffix string `json:"suffix"`
	Schema string `json:"types"`
}

type Field struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Endpoint    string `json:"endpoint"`
	Query       string `json:"query"`
}

type TypeConfig struct {
	Name   string  `json:"name"`
	Fields []Field `json:"fields"`
}

type Config struct {
	ConfigDirectory        string                  `json:"-"`
	DisableSchemaDownloads bool                    `json:"disable-schema-downloads"`
	EnabledSchemaStorage   bool                    `json:"enable-schema-storage"`
	Endpoints              map[string]EndpointInfo `json:"endpoints"`
	Types                  []TypeConfig            `json:"types"`
}

type endpoint struct {
	client func(request *graphql.Request) *graphql.Response
	schema *schema.Schema
	info   EndpointInfo
}

var validGraphQLIdentifierRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z_0-9]*$`)

func New(config Config) (*graphql.Engine, error) {

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
		c.HTTPClient = &http.Client{
			Transport: proxyTransport(0),
		}

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

	for _, typeConfig := range config.Types {
		object := root.Schema.Types[typeConfig.Name]
		if object == nil {
			object = &schema.Object{Name: typeConfig.Name}
		}
		if object, ok := object.(*schema.Object); ok {
			for _, fieldConfig := range typeConfig.Fields {
				if endpoint, ok := endpoints[fieldConfig.Endpoint]; ok {
					field := schema.Field{Name: fieldConfig.Name}
					if fieldConfig.Description != "" {
						field.Desc = schema.Description{Text: fieldConfig.Description}
					}
					err := mount(root, object.Name, field, fieldResolver, endpoint.schema, endpoint.client, fieldConfig.Query)
					if err != nil {
						return nil, err
					}
				} else {
					return nil, errors.New("invalid endpoint id: " + fieldConfig.Endpoint)
				}
			}
		} else {
			return nil, errors.Errorf("can only configure fields on OBJECT types: %s is a %s", typeConfig.Name, object.Kind())
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

var emptySelectionRegex = regexp.MustCompile(`{\s*}\s*$`)
var querySplitter = regexp.MustCompile(`[}\s]*$`)

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
				Context:   request.Context.GetContext(),
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

func getSelectedFields(upstreamSchema *schema.Schema, q *schema.QueryDocument, op *schema.Operation) ([]exec.FieldSelection, error) {
	onType := upstreamSchema.EntryPoints[op.Type]

	fsc := exec.FieldSelectionContext{
		Path:          []string{},
		Schema:        upstreamSchema,
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

		fsc.Path = append(fsc.Path, fields[0].Selection.Alias)
		fsc.OnType = fields[0].Field.Type
		selections = fields[0].Selection.Selections
	}
	return result, nil
}
