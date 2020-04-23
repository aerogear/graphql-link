package gateway

import (
	"net/http"
	"time"

	"github.com/chirino/graphql/resolvers"
)

type helpers map[string]interface{}

func (helpers) Login(rctx resolvers.ExecutionContext, args struct{ Token string }) string {
	ctx := rctx.GetContext()
	if r := ctx.Value("net/http.ResponseWriter"); r != nil {
		if r, ok := r.(http.ResponseWriter); ok {
			http.SetCookie(r, &http.Cookie{
				Name:    "Authorization",
				Value:   "Bearer " + args.Token,
				Path:    "/",
				Expires: time.Now().Add(1 * time.Hour),
			})
		}
	}
	return "ok"
}

func (helpers) Logout(rctx resolvers.ExecutionContext) string {
	ctx := rctx.GetContext()
	if r := ctx.Value("net/http.ResponseWriter"); r != nil {
		if r, ok := r.(http.ResponseWriter); ok {
			http.SetCookie(r, &http.Cookie{
				Name:    "Authorization",
				Value:   "",
				Path:    "/",
				Expires: time.Now().Add(-10000 * time.Hour),
			})
		}
	}
	return "ok"
}
