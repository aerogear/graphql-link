package main

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/3scale/kiper/pkg/queries"
	"github.com/chirino/graphql-gw/internal/dom"
	"github.com/open-policy-agent/opa/rego"
	"gopkg.in/yaml.v2"
)

func main() {

	// Load the input document.
	queries.RegisterThreeScaleQueries()
	queries.RegisterRateLimitQueries()

	ctx := context.Background()
	r := rego.New(
		rego.Package(`example`),
		rego.Query(`data.gateway`),
		rego.Module("gateway.rego", module))

	// Create a prepared query that can be evaluated.
	query, err := r.PrepareForEval(ctx)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		input, err := toEnvoyInput(request)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		rs, err := query.Eval(ctx, rego.EvalInput(input))
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		yaml.NewEncoder(writer).Encode(rs[0].Expressions[0].Value)
	})
	http.ListenAndServe("0.0.0.0:8080", nil)
}

func toEnvoyInput(r *http.Request) (interface{}, error) {

	parsedPath, err := parsePath(r.URL.Path)
	if err != nil {
		return nil, err
	}

	input := dom.Dom{
		"parsed_path":  parsedPath,
		"parsed_query": ValuesToMap(r.URL.Query()),
		// "parsed_body": ...,
		"attributes": dom.Dom{
			"source": dom.Dom{
				"address": GetRemoteIp(r),
				//"service": "",
				//"principal":"",
				//"certificate":"",
				//"labels": dom.Dom{},
			},
			//"destination": dom.Dom{
			//},
			"request": dom.Dom{
				"http": dom.Dom{
					"method":   strings.ToUpper(r.Method),
					"headers":  HeaderToMap(r.Header),
					"path":     r.URL.Path,
					"host":     r.URL.Host,
					"protocol": r.Proto,
					// "body":    ...,
				},
			},
		},
	}
	return input, nil
}

func ValuesToMap(from url.Values) (to map[string]string) {
	to = make(map[string]string, len(from))
	for k, _ := range from {
		to[k] = from.Get(k)
	}
	return
}

func HeaderToMap(from http.Header) (to map[string]string) {
	to = make(map[string]string, len(from))
	for k, _ := range from {
		to[strings.ToLower(k)] = from.Get(k)
	}
	return
}

func GetRemoteIp(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return ip
}

func parsePath(path string) ([]interface{}, error) {
	if len(path) == 0 {
		return []interface{}{}, nil
	}
	parts := strings.Split(path[1:], "/")
	for i := range parts {
		var err error
		parts[i], err = url.PathUnescape(parts[i])
		if err != nil {
			return nil, err
		}
	}
	sl := make([]interface{}, len(parts))
	for i := range sl {
		sl[i] = parts[i]
	}
	return sl, nil
}

const module = `
package gateway
import input.attributes.request as request

default drop_headers = []
default set_headers = []
default max_depth = 50
default max_parallelism = 10

headersThatMatch(allHeaders, re) = [key | val := allHeaders[key]; re_match(re, key)]
drop_headers = headersThatMatch(request.http.headers, "^sec.*")
`
