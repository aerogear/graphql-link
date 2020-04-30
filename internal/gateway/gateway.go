package gateway

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
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
	root.Resolver = resolvers.List(root.Resolver, upstreamDomResolverInstance, fieldResolver)

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
