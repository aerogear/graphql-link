package gateway

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql/schema"
)

type DataLoaders struct {
	started bool
	loaders map[string]*UpstreamDataLoader
}

type dataLoadersKey byte
const DataLoadersKey = dataLoadersKey(0)

type UpstreamDataLoader struct {
	ctx       context.Context
	upstream  *upstreamServer
	queryDocs []*schema.QueryDocument
	variables map[string]interface{}
	mergedDoc *schema.QueryDocument

	once     sync.Once
	response *graphql.Response
}

func (load *UpstreamDataLoader) resolution() (value reflect.Value, err error) {
	// concurrent call to Do will wait for the first call to finish..
	load.once.Do(func() {
		query := load.mergedDoc.String()
		load.response = load.upstream.client(&graphql.Request{
			Context:   load.ctx,
			Query:     query,
			Variables: load.variables,
		})
	})
	return reflect.Value{}, nil
}

func mergeQueryDocs(docs []*schema.QueryDocument) *schema.QueryDocument {
	toDoc := &schema.QueryDocument{}
	operations := map[schema.OperationType]*schema.Operation{}

	for _, d := range docs {
		fromOp := d.Operations[0]
		toOp := operations[fromOp.Type]
		if toOp == nil {
			operations[fromOp.Type] = fromOp
			toDoc.Operations = append(toDoc.Operations, fromOp)
		} else {
			for _, v := range fromOp.Vars {
				value := toOp.Vars.Get(v.Name)
				if value == nil {
					toOp.Vars = append(toOp.Vars, v)
				}
			}
			toOp.Selections = append(toOp.Selections, fromOp.Selections...)
			for _, fragment := range d.Fragments {
				if toDoc.Fragments.Get(fragment.Name) == nil {
					toDoc.Fragments = append(toDoc.Fragments, fragment)
				}
			}
		}
	}

	var counter int32 = 0
	for _, operation := range operations {
		operation.Selections = mergeQuerySelections(toDoc, operation.Selections, &counter)
	}

	return toDoc
}

func mergeQuerySelections(doc *schema.QueryDocument, from schema.SelectionList, counter *int32) schema.SelectionList {

	buf := &bytes.Buffer{}
	idx := map[string]schema.Selection{}
	result := schema.SelectionList{}

	for _, sel := range from {
		switch original := sel.(type) {
		case *schema.FieldSelection:
			buf.Reset()
			buf.WriteString(original.Name)
			original.Arguments.WriteTo(buf)
			original.Directives.WriteTo(buf)
			key := buf.String()

			if existing, ok := idx[key]; !ok {
				copy := *original
				result = append(result, &copy)
				idx[key] = &copy
				copy.Alias = fmt.Sprintf("f%x", *counter)
				original.Extension = copy.Alias
				*counter++
			} else {
				// Collapse dup field
				existing := existing.(*schema.FieldSelection)
				original.Extension = existing.Alias
				existing.Selections = append(existing.Selections, original.Selections...)
			}

		case *schema.InlineFragment:

			buf.Reset()
			buf.WriteString("... on ")
			original.On.WriteTo(buf)
			key := buf.String()

			if existing, ok := idx[key]; !ok {
				result = append(result, original)
				idx[key] = original
			} else {
				existing := existing.(*schema.InlineFragment)
				existing.Selections = mergeQuerySelections(doc, original.Selections, counter)
			}

		case *schema.FragmentSpread:

			buf.Reset()
			buf.WriteString("...")
			buf.WriteString(original.Name)
			key := buf.String()

			if _, ok := idx[key]; !ok {
				result = append(result, original)
				idx[key] = original
			}
		}
	}

	for _, sel := range result {
		switch sel := sel.(type) {
		case *schema.FieldSelection:
			sel.Selections = mergeQuerySelections(doc, sel.Selections, counter)
		case *schema.InlineFragment:
			sel.Selections = mergeQuerySelections(doc, sel.Selections, counter)
		}
	}
	return result
}
