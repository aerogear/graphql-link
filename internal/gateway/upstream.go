package gateway

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type Upstream struct {
	Type string `json:"type"`
}

func (a *Upstream) GetUpstream() *Upstream {
	return a
}

type upstreamGetter interface {
	GetUpstream() *Upstream
}

type UpstreamWrapper struct {
	Upstream upstreamGetter `json:"-"`
}

func (h *UpstreamWrapper) UnmarshalJSON(b []byte) error {
	raw := json.RawMessage{}
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	a := Upstream{}
	err = json.Unmarshal(raw, &a)
	if err != nil {
		return err
	}

	var upstream upstreamGetter
	switch a.Type {
	case "":
		upstream = &GraphQLUpstream{}
	case "graphql":
		upstream = &GraphQLUpstream{}
	default:
		return errors.New("invalid action type")
	}

	err = json.Unmarshal(raw, upstream)
	if err != nil {
		return err
	}

	h.Upstream = upstream
	return nil
}

func (f *UpstreamWrapper) MarshalJSON() ([]byte, error) {
	if f.Upstream != nil {
		typeValue := ""
		switch f.Upstream.(type) {
		case *GraphQLUpstream:
			typeValue = "graphql"
		}
		f.Upstream.GetUpstream().Type = typeValue
	}
	return json.Marshal(f.Upstream)
}

type GraphQLUpstream struct {
	Upstream
	Prefix string `json:"prefix"`
	Suffix string `json:"suffix"`
	URL    string `json:"url"`
	Schema string `json:"types"`
}
