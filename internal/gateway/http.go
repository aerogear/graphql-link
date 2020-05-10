package gateway

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql/graphiql"
	"github.com/chirino/graphql/httpgql"
	"github.com/pkg/errors"
)

func CreateHttpHandler(f graphql.ServeGraphQLStreamFunc) http.Handler {
	return &httpgql.Handler{ServeGraphQLStream: f}
}

type proxyTransport byte

func (p proxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {

	ctx := req.Context()
	if ctx != nil {

		value := ctx.Value("*net/http.Request")
		if value != nil {
			originalRequest := value.(*http.Request)

			if clientIP, _, err := net.SplitHostPort(originalRequest.RemoteAddr); err == nil {
				if prior, ok := originalRequest.Header["X-Forwarded-For"]; ok {
					clientIP = strings.Join(prior, ", ") + ", " + clientIP
				}
				req.Header.Set("X-Forwarded-For", clientIP)
			}

			if _, ok := originalRequest.Header["X-Forwarded-Host"]; !ok {
				if host := originalRequest.Header.Get("Host"); host != "" {
					req.Header.Set("X-Forwarded-Host", host)
				}
			}

			if _, ok := originalRequest.Header["X-Forwarded-Proto"]; !ok {
				if originalRequest.TLS != nil {
					req.Header.Set("X-Forwarded-Proto", "https")
				} else {
					req.Header.Set("X-Forwarded-Proto", "http")
				}
			}
		}

	}
	return http.DefaultTransport.RoundTrip(req)
}

func StartServer(host string, port uint16, engine *graphql.Engine, log *log.Logger) (*httptest.Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		switch host {
		case "localhost":
			fallthrough
		case "127.0.0.1":
			host = "[::1]"
			if l, err = net.Listen("tcp6", fmt.Sprintf("%s:%d", host, port)); err != nil {
				panic(fmt.Sprintf("httptest: failed to listen on a port: %v", err))
			}
		default:
			return nil, errors.Wrap(err, "httptest: failed to listen on a port")
		}
	}

	mux := http.NewServeMux()
	mux.Handle("/graphql", &httpgql.Handler{ServeGraphQLStream: engine.ServeGraphQLStream})
	endpoint := fmt.Sprintf("http://%s/graphql", l.Addr())
	mux.Handle("/", graphiql.New(endpoint, true))
	ts := &httptest.Server{
		Listener: l,
		Config:   &http.Server{Handler: mux},
	}
	log.Printf("GraphQL endpoint running at %s", endpoint)
	log.Printf("GraphQL UI running at http://%s", l.Addr())
	ts.Start()
	return ts, nil
}
