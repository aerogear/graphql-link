package gateway

import (
	"encoding/json"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql/resolvers"
	"github.com/chirino/graphql/schema"
	"github.com/pkg/errors"
)

type actionRunner struct {
	Gateway   *graphql.Engine
	Endpoints map[string]*upstreamServer
	Resolver  resolvers.TypeAndFieldResolver
	Type      *schema.Object
}

type Action struct {
	Type string `json:"type"`
}

func (a *Action) GetAction() *Action {
	return a
}

type actionGetter interface {
	GetAction() *Action
}

type ActionWrapper struct {
	Action actionGetter `json:"-"`
}

func (h *ActionWrapper) UnmarshalJSON(b []byte) error {
	raw := json.RawMessage{}
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	a := Action{}
	err = json.Unmarshal(raw, &a)
	if err != nil {
		return err
	}

	var action actionGetter
	switch a.Type {
	case "mount":
		action = &Mount{}
	case "rename":
		action = &Rename{}
	case "remove":
		action = &Remove{}
	case "link":
		action = &Link{}
	default:
		return errors.New("invalid action type")
	}

	err = json.Unmarshal(raw, action)
	if err != nil {
		return err
	}

	h.Action = action
	return nil
}

func (f *ActionWrapper) MarshalJSON() ([]byte, error) {
	if f.Action != nil {
		typeValue := ""
		switch f.Action.(type) {
		case *Mount:
			typeValue = "mount"
		case *Rename:
			typeValue = "rename"
		case *Link:
			typeValue = "link"
		case *Remove:
			typeValue = "remove"
		}
		f.Action.GetAction().Type = typeValue
	}
	return json.Marshal(f.Action)
}
