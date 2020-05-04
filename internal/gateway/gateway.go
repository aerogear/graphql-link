package gateway

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql/exec"
	"github.com/chirino/graphql/httpgql"
	"github.com/chirino/graphql/qerrors"
	"github.com/chirino/graphql/resolvers"
	"github.com/chirino/graphql/schema"
	"github.com/pkg/errors"
)

type TypeConfig struct {
	Name    string          `json:"name"`
	Actions []ActionWrapper `json:"actions"`
}

type Config struct {
	ConfigDirectory        string                     `json:"-"`
	DisableSchemaDownloads bool                       `json:"disable-schema-downloads"`
	EnabledSchemaStorage   bool                       `json:"enable-schema-storage"`
	Upstreams              map[string]UpstreamWrapper `json:"upstreams"`
	Types                  []TypeConfig               `json:"types"`
}

type upstreamServer struct {
	client func(request *graphql.Request) *graphql.Response
	schema *schema.Schema
	info   GraphQLUpstream
}

var validGraphQLIdentifierRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z_0-9]*$`)

func New(config Config) (*graphql.Engine, error) {

	if config.ConfigDirectory == "" {
		config.ConfigDirectory = "."
	}
	if config.EnabledSchemaStorage {
		os.MkdirAll(filepath.Join(config.ConfigDirectory, "upstreams"), 0755)
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

	upstreams := map[string]*upstreamServer{}

	for eid, upstream := range config.Upstreams {
		switch upstream := upstream.Upstream.(type) {
		case *GraphQLUpstream:
			c := httpgql.NewClient(upstream.URL)
			c.HTTPClient = &http.Client{
				Transport: proxyTransport(0),
			}
			upstreams[eid] = &upstreamServer{
				info:   *upstream,
				client: c.ServeGraphQL,
			}
		default:
			panic("invalid upstream type")
		}
	}

	for eid, upstream := range upstreams {
		s, err := loadEndpointSchema(config, eid, upstream)
		if err != nil {
			return nil, err
		}

		if upstream.info.Prefix != "" {
			s.RenameTypes(func(x string) string { return upstream.info.Prefix + x })
		}
		if upstream.info.Suffix != "" {
			s.RenameTypes(func(x string) string { return x + upstream.info.Suffix })
		}
		upstreams[eid].schema = s
	}

	actionRunner := actionRunner{
		Gateway:   root,
		Endpoints: upstreams,
		Resolver:  fieldResolver,
	}
	for _, typeConfig := range config.Types {
		object := root.Schema.Types[typeConfig.Name]
		if object == nil {
			object = &schema.Object{Name: typeConfig.Name}
		}
		if object, ok := object.(*schema.Object); ok {
			actionRunner.Type = object
			for _, action := range typeConfig.Actions {
				switch action := action.Action.(type) {
				case *Mount:
					err := actionRunner.mount(action)
					if err != nil {
						return nil, err
					}
				case *Rename:

				}

			}
		} else {
			return nil, errors.Errorf("can only configure fields on OBJECT types: %s is a %s", typeConfig.Name, object.Kind())
		}
	}
	return root, nil
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
