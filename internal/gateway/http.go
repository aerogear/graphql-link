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

			for k, headers := range originalRequest.Header {
				switch k {

				// Hop-by-hop headers... Don't forward these.
				// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
				case "Connection":
				case "Keep-Alive":
				case "Proxy-Authenticate":
				case "Proxy-Authorization":
				case "Te":
				case "Trailers":
				case "Transfer-Encoding":
				case "Upgrade":

				// Skip these headers which could affect our connection
				// to the upstream:
				case "Accept-Encoding":
				case "Sec-Websocket-Version":
				case "Sec-Websocket-Protocol":
				case "Sec-Websocket-Extensions":
				case "Sec-Websocket-Key":
				default:
					// Copy over any other headers..
					for _, header := range headers {
						req.Header.Add(k, header)
					}
				}
			}

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

	mux := http.NewServeMux()
	server, err := StartHttpListener(fmt.Sprintf("%s:%d", host, port), mux)
	if err != nil {
		return nil, err
	}

	mux.Handle("/graphql", &httpgql.Handler{ServeGraphQLStream: engine.ServeGraphQLStream})
	endpoint := fmt.Sprintf("%s/graphql", server.URL)
	mux.Handle("/", graphiql.New(endpoint, true))

	log.Printf("GraphQL endpoint running at %s", endpoint)
	log.Printf("GraphQL UI running at %s", server.URL)
	return server, nil
}

func StartHttpListener(listen string, handler http.Handler) (*httptest.Server, error) {
	host, port, err := net.SplitHostPort(listen)
	if err != nil {
		return nil, err
	}

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		switch host {
		case "localhost":
			fallthrough
		case "127.0.0.1":
			host = "[::1]"
			if l, err = net.Listen("tcp6", fmt.Sprintf("%s:%s", host, port)); err != nil {
				return nil, errors.Wrap(err, "failed to listen on the port")
			}
		default:
			return nil, errors.Wrap(err, "failed to listen on the port")
		}
	}
	ts := &httptest.Server{
		Listener: l,
		Config:   &http.Server{Handler: handler},
	}
	ts.Start()
	return ts, nil
}
