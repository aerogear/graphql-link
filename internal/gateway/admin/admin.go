package admin

import (
	"io/ioutil"
	"net/http"

	"github.com/aerogear/graphql-link/internal/cmd/config"
	ghodssyaml "github.com/ghodss/yaml"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"gopkg.in/yaml.v2"
)

func CreateHttpHandler() http.Handler {
	admin := admin{}
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Route("/config", func(r chi.Router) {
		r.Get("/", admin.GetConfig)
		r.Post("/", admin.UpdateConfig)
	})

	return r
}

type admin struct {
	configFile string
}

func (h *admin) UpdateConfig(w http.ResponseWriter, r *http.Request) {

	// JSON -> YAML first...
	dom := map[string]interface{}{}
	err := render.DecodeJSON(r.Body, &dom)
	if err != nil {
		render.Render(w, r, RenderErr(err, http.StatusBadRequest, "Bad Request"))
		return
	}
	data, err := yaml.Marshal(dom)
	if err != nil {
		render.Render(w, r, RenderErr(err, http.StatusBadRequest, "Bad Request"))
		return
	}

	// data is now in YAML format... now check it decodes into the Config struct...
	c := config.Config{}
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		render.Render(w, r, RenderErr(err, http.StatusBadRequest, "Bad Request"))
		return
	}
	err = config.Store(c)
	if err != nil {
		render.Render(w, r, RenderErr(err, http.StatusInternalServerError, "Internal Server Error"))
		return
	}
	render.NoContent(w, r)
}

func (h *admin) GetConfig(w http.ResponseWriter, r *http.Request) {
	// YAML to JSON conversion...
	data, err := ioutil.ReadFile(config.File)
	dom := map[string]interface{}{}
	err = ghodssyaml.Unmarshal(data, &dom)
	if err != nil {
		render.Render(w, r, RenderErr(err, http.StatusInternalServerError, "Internal Server Error"))
		return
	}
	render.JSON(w, r, dom)
}

func RenderErr(err error, code int, status string) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: code,
		StatusText:     status,
		ErrorText:      err.Error(),
	}
}

type ErrResponse struct {
	Err            error  `json:"-"`               // low-level runtime error
	HTTPStatusCode int    `json:"-"`               // http response status code
	StatusText     string `json:"status"`          // user-level status message
	ErrorText      string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}
