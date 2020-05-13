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

	return func() (reflect.Value, error) {
		// In case we need to debug...
		//json, _ := json.Marshal(parentValue.Interface())
		//fmt.Println(string(json))

		selection := request.Selection
		field := reflect.ValueOf(selection.Extension)
		value := parentValue.MapIndex(field)
		if !value.IsValid() {
			return reflect.Zero(parentValue.Type().Elem()), nil
		}
		if value.Interface() == nil {
			return value, nil
		}
		value = reflect.ValueOf(value.Interface())
		return value, nil
	}

}
