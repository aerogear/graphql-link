package gateway

import (
	"reflect"

	"github.com/chirino/graphql/resolvers"
)

type upstreamDomResolver byte

var upstreamDomResolverInstance = upstreamDomResolver(0)

// the upstream results already have the results using the requested field aliases.. so
// when we request a given field name, we should actually use the alias name to look up values
// in the json maps.
func (r upstreamDomResolver) Resolve(request *resolvers.ResolveRequest, next resolvers.Resolution) resolvers.Resolution {
	if request.Context.Value(upstreamDomResolverInstance) == nil {
		return next
	}

	// This is basically exactly like
	parentValue := resolvers.Dereference(request.Parent)
	if parentValue.Kind() != reflect.Map || parentValue.Type().Key().Kind() != reflect.String {
		return next
	}

	selection := request.Selection

	//In case we need to debug...
	//dump, _ := json.Marshal(parentValue.Interface())
	//fmt.Println(string(dump))

	var field reflect.Value
	if selection.Extension != nil {
		field = reflect.ValueOf(selection.Extension)
	} else {
		// TODO: try to eliminate this case...
		field = reflect.ValueOf(selection.Alias)
	}
	value := parentValue.MapIndex(field)
	if !value.IsValid() {
		value = reflect.Zero(parentValue.Type().Elem())
	}

	//dump, _ = json.Marshal(value.Interface())
	//fmt.Println(string(dump))
	// value = reflect.ValueOf(value.Interface())

	return func() (reflect.Value, error) {
		return value, nil
	}
}
