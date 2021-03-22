package gateway

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql-4-apis/pkg/apis"
	"github.com/chirino/graphql/exec"
	"github.com/chirino/graphql/httpgql"
	"github.com/chirino/graphql/resolvers"
	"github.com/chirino/graphql/schema"
	qerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/pkg/errors"
)

type TypeConfig struct {
	Name    string          `yaml:"name,omitempty"`
	Actions []ActionWrapper `yaml:"actions,omitempty"`
}

type SchemaConfig struct {
	Query        string `yaml:"query,omitempty"`
	Mutation     string `yaml:"mutation,omitempty"`
	Subscription string `yaml:"subscription,omitempty"`
}

type PolicyAgentConfig struct {
	Address string `yaml:"address,omitempty"`
	// InsecureClient allows connections to servers that do not have a valid TLS certificate.
	InsecureClient bool `yaml:"insecure-client,omitempty",json:"insecure-client,omitempty"`
}

type Config struct {
	WorkDirectory          string                     `yaml:"-"`
	Log                    *log.Logger                `yaml:"-"`
	DisableSchemaDownloads bool                       `yaml:"disable-schema-downloads,omitempty"`
	EnabledSchemaStorage   bool                       `yaml:"enable-schema-storage,omitempty"`
	Upstreams              map[string]UpstreamWrapper `yaml:"upstreams"`
	Schema                 *SchemaConfig              `yaml:"schema,omitempty"`
	Types                  []TypeConfig               `yaml:"types"`
	PolicyAgent            PolicyAgentConfig          `yaml:"policy-agent"`
}

var validGraphQLIdentifierRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z_0-9]*$`)

type Gateway struct {
	*graphql.Engine
	onClose []func()
}

type ModifiedQueryError struct {
	*qerrors.QueryError
	stack errors.StackTrace
}

func (gw *Gateway) Close() {
	funcs := gw.onClose
	gw.onClose = nil
	for _, f := range funcs {
		f()
	}
}

func New(config Config) (*Gateway, error) {
	if config.Log == nil {
		config.Log = NoLog
	}
	if config.WorkDirectory == "" {
		config.WorkDirectory = "."
	}
	if config.EnabledSchemaStorage {
		os.MkdirAll(filepath.Join(config.WorkDirectory, "upstreams"), 0755)
	}

	fieldResolver := resolvers.TypeAndFieldResolver{}
	gateway := &Gateway{Engine: graphql.New()}
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
	if config.Schema != nil {
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
	}

	if err != nil {
		panic(err)
	}
	gateway.Resolver = resolvers.List(gateway.Resolver, upstreamDomResolverInstance, fieldResolver)

	upstreams := createUpstreams(config)

	for eid, upstream := range upstreams {

		original, err := loadEndpointSchema(config, upstream)
		if err != nil {
			log.Printf("%v", err)
			continue
			//return nil, err
		}

		upstreams[eid].RenameTypes(original)

	}

	actionRunner := actionRunner{
		Gateway:   gateway.Engine,
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
						log.Printf("%v", err)
						continue
						//return nil, err
					}
				case *Link:
					err := actionRunner.link(action)
					if err != nil {
						log.Printf("%v", err)
						continue
						//return nil, err
					}
				case *Rename:
					err := actionRunner.rename(action)
					if err != nil {
						log.Printf("%v", err)
						continue
						//return nil, err
					}
				case *Remove:
					err := actionRunner.remove(action)
					if err != nil {
						log.Printf("%v", err)
						continue
						//return nil, err
					}
				}

			}
		} else {
			return nil, errors.Errorf("can only configure fields on OBJECT types: %s is a %s", typeConfig.Name, object.Kind())
		}
	}

	err = initPolicyAgent(config, gateway)
	if err != nil {
		return nil, err
	}

	return gateway, nil
}

func createUpstreams(config Config) map[string]*upstreamServer {
	upstreams := map[string]*upstreamServer{}

	for upstreamId, upstream := range config.Upstreams {
		switch upstream := upstream.Upstream.(type) {
		case *GraphQLUpstream:
			u, err := CreateGraphQLUpstreamServer(upstreamId, upstream)
			if err != nil {
				config.Log.Printf("upstream '%s' disabled: %v", upstreamId, err)
				continue
			}
			upstreams[upstreamId] = u
		case *OpenApiUpstream:

			u, err := CreateOpenAPIUpstreamServer(upstreamId, upstream, config)
			if err != nil {
				config.Log.Printf("upstream '%s' disabled: %v", upstreamId, err)
				continue
			}
			upstreams[upstreamId] = u

		default:
			panic("invalid upstream type")
		}
	}
	return upstreams
}

func CreateOpenAPIUpstreamServer(id string, upstream *OpenApiUpstream, config Config) (*upstreamServer, error) {
	engine, err := apis.CreateGatewayEngine(apis.Config{
		Openapi: upstream.Openapi,
		APIBase: upstream.APIBase,
		Log:     SimpleLog,
	})
	if err != nil {
		return nil, err
	}
	return &upstreamServer{
		id:                         id,
		Client:                     engine.ServeGraphQL,
		subscriptionClient:         engine.ServeGraphQLStream,
		originalNames:              map[string]schema.NamedType{},
		gatewayToUpstreamTypeNames: map[string]string{},
		info: UpstreamInfo{
			URL:     upstream.Openapi.URL, // todo.. replace with the resolved API address
			Prefix:  upstream.Prefix,
			Suffix:  upstream.Suffix,
			Headers: upstream.Headers,
		},
		Schema: nil,
	}, nil
}

func CreateGraphQLUpstreamServer(id string, upstream *GraphQLUpstream) (*upstreamServer, error) {
	if upstream.URL == "" {
		return nil, errors.New("url is not configured")
	}

	_, err := url.Parse(upstream.URL)
	if err != nil {
		return nil, err
	}

	c := httpgql.NewClient(upstream.URL)
	info := UpstreamInfo{
		URL:     upstream.URL,
		Prefix:  upstream.Prefix,
		Suffix:  upstream.Suffix,
		Schema:  upstream.Schema,
		Headers: upstream.Headers,
	}
	c.HTTPClient = &http.Client{
		Transport: &info,
	}
	return &upstreamServer{
		id:                         id,
		Client:                     c.ServeGraphQL,
		subscriptionClient:         c.ServeGraphQLStream,
		originalNames:              map[string]schema.NamedType{},
		gatewayToUpstreamTypeNames: map[string]string{},
		info:                       info,
		Schema:                     nil,
	}, nil
}

func HaveUpstreamSchemaChanged(config Config) (bool, error) {
	if config.DisableSchemaDownloads || !config.EnabledSchemaStorage {
		return false, nil
	}
	upstreams := createUpstreams(config)
	for eid, upstream := range upstreams {

		// Load the old stored schema.
		upstreamSchemaFile := filepath.Join(config.WorkDirectory, "upstreams", eid+".graphql")
		data, err := ioutil.ReadFile(upstreamSchemaFile)
		if err != nil {
			return false, err
		}

		s, err := downloadSchema(config, upstream)
		if err != nil {
			continue // ignore download errors... they could be transient..
		}
		if s.String() != string(data) {
			return true, nil
		}
	}
	return false, nil
}

func (e *ModifiedQueryError) WithLocations(locations ...qerrors.Location) *ModifiedQueryError {
	e.Locations = locations
	return e
}

func (e *ModifiedQueryError) WithStack() *ModifiedQueryError {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	f := make([]errors.Frame, n)
	for i := 0; i < n; i++ {
		f[i] = errors.Frame(pcs[i])
	}

	e.stack = f
	return e
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
			loc := qerrors.Location(firstSelection.Location())
			err := &ModifiedQueryError{QueryError: qerrors.Errorf("No fields selected")}
			return nil, err.WithLocations(loc).WithStack()
		}
		if len(fields) > 1 {
			loc := qerrors.Location(firstSelection.Location())
			err := &ModifiedQueryError{QueryError: qerrors.Errorf("please only select one field")}
			return nil, err.WithLocations(loc).WithStack()
		}
		result = append(result, fields[0])

		fsc.Path = append(fsc.Path, fields[0].Selection.Alias)
		fsc.OnType = fields[0].Field.Type
		selections = fields[0].Selection.Selections
	}
	return result, nil
}
