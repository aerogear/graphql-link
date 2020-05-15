package gateway

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql/resolvers"
	"github.com/chirino/graphql/schema"
	"github.com/pkg/errors"
)

type upstreamServer struct {
	id                         string
	client                     func(request *graphql.Request) *graphql.Response
	subscriptionClient         func(request *graphql.Request) graphql.ResponseStream
	originalNames              map[string]schema.NamedType
	gatewayToUpstreamTypeNames map[string]string
	schema                     *schema.Schema
	originalSchema             *schema.Schema
	info                       GraphQLUpstream
}

func (u *upstreamServer) ToUpstreamInputValueList(from schema.InputValueList) schema.InputValueList {
	joinedOpVars := schema.InputValueList{}
	for _, v := range from {
		c := *v
		c.Type = u.ToUpstreamType(c.Type)
		joinedOpVars = append(joinedOpVars, &c)
	}
	return joinedOpVars
}

func (u *upstreamServer) ToUpstreamType(t schema.Type) schema.Type {
	switch t := t.(type) {
	case *schema.NonNull:
		return &schema.NonNull{OfType: u.ToUpstreamType(t.OfType)}
	case *schema.List:
		return &schema.List{OfType: u.ToUpstreamType(t.OfType)}
	case *schema.TypeName:
		name := t.Name
		name = u.gatewayToUpstreamTypeNames[name]
		return &schema.TypeName{
			Name: name,
		}
	case schema.NamedType:
		name := t.TypeName()
		name = u.gatewayToUpstreamTypeNames[name]
		return &schema.TypeName{
			Name: name,
		}
	}
	return t
}

func getUpstreamValue(ctx context.Context, result *graphql.Response, doc *schema.QueryDocument, selectionPath []schema.Selection) (reflect.Value, error) {
	if len(result.Errors) > 0 {
		return reflect.Value{}, result.Error()
	}

	data := map[string]interface{}{}
	err := json.Unmarshal(result.Data, &data)
	if err != nil {
		return reflect.Value{}, err
	}
	var v interface{} = data

	for _, sel := range selectionPath {
		switch sel := sel.(type) {
		case *schema.FieldSelection:
			if m, ok := v.(map[string]interface{}); ok {
				v = m[sel.Extension.(string)]
			} else {
				return reflect.Value{}, errors.Errorf("expected upstream field not found: %s", sel.Name)
			}
		}
	}

	// This enables the upstreamDomResolverInstance for all child fields of this result.
	// needed to property handle field aliases.
	return reflect.ValueOf(resolvers.ValueWithContext{
		Value:   reflect.ValueOf(v),
		Context: context.WithValue(ctx, upstreamDomResolverInstance, true),
	}), nil
	//return reflect.ValueOf(v), nil
}
