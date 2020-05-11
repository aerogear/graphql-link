package dom

import (
	"encoding/json"
	"fmt"
	"github.com/chirino/graphql/text"
)

type Dom map[string]interface{}

func New(args...interface{}) Dom {
	return Dom{}.Set(args...)
}

func Parse(jsonText string) (Dom) {
	dom := Dom{}
	err := json.Unmarshal([]byte(jsonText), &dom)
	if err != nil {
		panic(fmt.Sprintf("%+v\n%s\n", err, text.BulletIndent("json: ", jsonText)))
	}
	return dom
}

func (d Dom) Get(path ...string) (interface{}, bool) {
	var x interface{} = d
	found := false
	for _, p := range path {
		switch y := x.(type) {
		case Dom:
			x, found = y[p]
			if !found {
				return nil, false
			}
			if y, ok := x.(map[string]interface{}); ok {
				x = Dom(y)
			}
		default:
			return nil, false
		}
	}
	return x, true
}

func (d Dom) GetString(path ...string) *string {
	if v, ok := d.Get(path...); ok {
		switch v := v.(type) {
		case string:
			return &v
		case *string:
			return v
		}
	}
	return nil
}

func (d Dom) GetDom(path ...string) *Dom {
	if v, ok := d.Get(path...); ok {
		switch v := v.(type) {
		case Dom:
			return &v
		}
	}
	return nil
}

func (d Dom) String() string {
	data, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (d Dom) Set(args...interface{}) Dom {
	for i := 0; i+1 < len(args); i += 2 {
		d[args[i].(string)] = args[i+1]
	}
	return d
}
