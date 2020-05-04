package gateway

import (
	"net"
	"net/http"
	"strings"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql/httpgql"
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
