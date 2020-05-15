package gateway

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
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

type SchemaConfig struct {
	Query        string `json:"query"`
	Mutation     string `json:"mutation"`
	Subscription string `json:"subscription"`
}

type Config struct {
	ConfigDirectory        string                     `json:"-"`
	Log                    *log.Logger                `json:"-"`
	DisableSchemaDownloads bool                       `json:"disable-schema-downloads"`
	EnabledSchemaStorage   bool                       `json:"enable-schema-storage"`
	Upstreams              map[string]UpstreamWrapper `json:"upstreams"`
	Schema                 SchemaConfig
	Types                  []TypeConfig `json:"types"`
}

var validGraphQLIdentifierRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z_0-9]*$`)

func New(config Config) (*graphql.Engine, error) {
	if config.Log == nil {
		config.Log = NoLog
	}
	if config.ConfigDirectory == "" {
		config.ConfigDirectory = "."
	}
	if config.EnabledSchemaStorage {
		os.MkdirAll(filepath.Join(config.ConfigDirectory, "upstreams"), 0755)
	}

	fieldResolver := resolvers.TypeAndFieldResolver{}
	gateway := graphql.New()
	err := gateway.Schema.Parse(`
schema {
    query: Query
    mutation: Mutation
    subscription: Subscription
}
type Query {}
type Mutation {}
type Subscription {}
`)

	// Use the OnRequestHook to add UpstreamLoads to the query context
	gateway.OnRequestHook = func(r *graphql.Request, doc *schema.QueryDocument, op *schema.Operation) error {
		r.Context = context.WithValue(r.GetContext(), DataLoadersKey, &DataLoaders{
			loaders: map[string]*UpstreamDataLoader{},
		})
		return nil
	}
	defaultTryCast := gateway.TryCast
	gateway.TryCast = func(value reflect.Value, toType string) (reflect.Value, bool) {

		// First try the default cast strategy...
		cast, ok := defaultTryCast(value, toType)
		if ok {
			return cast, ok
		}

		// fallback to using "__typename" data
		dv := resolvers.Dereference(value)
		if m, ok := dv.Interface().(map[string]interface{}); ok {
			if tn, ok := m["t"]; ok {
				if tn == toType {
					return value, true
				}
			}
		}
		return value, false
	}

	// To support configuring the names of the root types.
	gateway.Schema.RenameTypes(func(n string) string {
		switch n {
		case "Query":
			if config.Schema.Query != "" {
				return config.Schema.Query
			}
		case "Mutation":
			if config.Schema.Mutation != "" {
				return config.Schema.Mutation
			}
		case "Subscription":
			if config.Schema.Subscription != "" {
				return config.Schema.Subscription
			}
		}
		return n
	})

	if err != nil {
		panic(err)
	}
	gateway.Resolver = resolvers.List(gateway.Resolver, upstreamDomResolverInstance, fieldResolver)

	upstreams := map[string]*upstreamServer{}

	for upstreamId, upstream := range config.Upstreams {
		switch upstream := upstream.Upstream.(type) {
		case *GraphQLUpstream:
			c := httpgql.NewClient(upstream.URL)
			c.HTTPClient = &http.Client{
				Transport: proxyTransport(0),
			}
			upstreams[upstreamId] = &upstreamServer{
				id:                         upstreamId,
				client:                     c.ServeGraphQL,
				subscriptionClient:         c.ServeGraphQLStream,
				originalNames:              map[string]schema.NamedType{},
				gatewayToUpstreamTypeNames: map[string]string{},
				info:                       *upstream,
				schema:                     nil,
			}
		default:
			panic("invalid upstream type")
		}
	}

	for eid, upstream := range upstreams {
		original, err := loadEndpointSchema(config, eid, upstream)

		// TODO: implement schema.DeepCopy()
		merged := schema.New()
		err = merged.Parse(original.String())
		if err != nil {
			panic(err)
		}

		if err != nil {
			return nil, err
		}

		for k, v := range merged.Types {
			upstream.originalNames[k] = v
		}
		if upstream.info.Prefix != "" {
			merged.RenameTypes(func(x string) string { return upstream.info.Prefix + x })
		}
		if upstream.info.Suffix != "" {
			merged.RenameTypes(func(x string) string { return x + upstream.info.Suffix })
		}
		upstreams[eid].schema = merged
		upstreams[eid].originalSchema = original
		for n, t := range upstream.originalNames {
			upstream.gatewayToUpstreamTypeNames[t.TypeName()] = n
		}
	}

	actionRunner := actionRunner{
		Gateway:   gateway,
		Endpoints: upstreams,
		Resolver:  fieldResolver,
	}
	for _, typeConfig := range config.Types {
		object := gateway.Schema.Types[typeConfig.Name]
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
					err := actionRunner.rename(action)
					if err != nil {
						return nil, err
					}
				}

			}
		} else {
			return nil, errors.Errorf("can only configure fields on OBJECT types: %s is a %s", typeConfig.Name, object.Kind())
		}
	}
	return gateway, nil
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
