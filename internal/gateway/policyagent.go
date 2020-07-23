package gateway

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql-gw/internal/gateway/policyagent/proto"
	"github.com/chirino/graphql/exec"
	"github.com/chirino/graphql/qerrors"
	"github.com/chirino/graphql/schema"
	"google.golang.org/grpc"
)

var checkResponseFieldsKey = "checkResponseFieldsKey"

func getFieldPolicies(ctx context.Context) []*proto.GraphQLFieldResponse {
	value := ctx.Value(checkResponseFieldsKey)
	if value == nil {
		return nil
	}
	return value.([]*proto.GraphQLFieldResponse)
}

func initPolicyAgent(config Config, gateway *Gateway) error {
	if config.PolicyAgent.Address != "" {
		opts := []grpc.DialOption{}
		if config.PolicyAgent.InsecureClient {
			opts = append(opts, grpc.WithInsecure())
		}
		c, err := grpc.Dial(config.PolicyAgent.Address, opts...)
		if err != nil {
			return err
		}
		gateway.onClose = append(gateway.onClose, func() {
			c.Close()
		})
		policyAgentClient := proto.NewPolicyAgentClient(c)

		originalValidate := gateway.Validate
		originalOnRequestHook := gateway.OnRequestHook
		gateway.OnRequestHook = func(request *graphql.Request, doc *schema.QueryDocument, op *schema.Operation) error {

			err := originalOnRequestHook(request, doc, op)
			if err != nil {
				return err
			}

			err = originalValidate(doc, gateway.MaxDepth)
			if err != nil {
				return err
			}

			fields := toPolicyCheckFields(gateway.Schema, doc, op)
			if len(fields) == 0 {
				return nil
			}

			r := getHttpRequest(request.Context)
			checkRequest := &proto.CheckRequest{
				Source: &proto.Source{
					Address: GetRemoteIp(r),
				},
				Http: &proto.Http{
					Host:     r.URL.Host,
					Method:   strings.ToUpper(r.Method),
					Path:     r.URL.Path,
					Protocol: r.Proto,
					Headers:  toPolicyCheckHeaders(r.Header),
				},
				Graphql: &proto.GraphQL{
					Fields: fields,
				},
			}

			checkResponse, err := policyAgentClient.Check(request.GetContext(), checkRequest)
			if err != nil {
				return err
			}

			if len(checkResponse.ValidationError) > 0 {
				x := qerrors.ErrorList{}
				for _, e := range checkResponse.ValidationError {
					x = append(x, qerrors.New(e))
				}
				return x.Error()
			}

			if len(checkResponse.Fields) > 0 {
				request.Context = context.WithValue(request.GetContext(), checkResponseFieldsKey, checkResponse.Fields)
			}
			return nil
		}
		gateway.Validate = func(doc *schema.QueryDocument, maxDepth int) error {
			return nil
		}
	}
	return nil
}

func toPolicyCheckFields(gwschema *schema.Schema, doc *schema.QueryDocument, op *schema.Operation) []*proto.GraphQLField {
	path := []string{string(op.Type)}
	onType := gwschema.EntryPoints[op.Type]
	return collectPolicyCheckFields(gwschema, onType, doc, path, op, nil)
}

func collectPolicyCheckFields(gwschema *schema.Schema, onType schema.Type,
	doc *schema.QueryDocument, path []string, s schema.Selection, to []*proto.GraphQLField) []*proto.GraphQLField {

	fsc := exec.FieldSelectionContext{
		Path:          path,
		Schema:        gwschema,
		QueryDocument: doc,
		OnType:        onType,
	}
	fields, errs := fsc.Apply(s.GetSelections(doc))
	if errs != nil {
		return to
	}
	for _, f := range fields {
		to = append(to, &proto.GraphQLField{
			Path:       strings.Join(append(path, f.Field.Name), "/"),
			ParentType: f.OnType.String(),
			FieldType:  f.Field.Type.String(),
			Args:       "{}", // TODO
		})
		if len(f.Selection.Selections) != 0 {
			to = collectPolicyCheckFields(gwschema, f.Field.Type, doc, append(path, f.Field.Name), f.Selection, to)
		}
	}
	return to
}

func toPolicyCheckHeaders(header http.Header) []*proto.Header {
	rc := []*proto.Header{}
	for k, h := range header {
		for _, v := range h {
			rc = append(rc, &proto.Header{Name: k, Value: v})
		}
	}
	return rc
}

func GetRemoteIp(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return ip
}
