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
		}
		for _, fragment := range d.Fragments {
			if toDoc.Fragments.Get(fragment.Name) == nil {
				toDoc.Fragments = append(toDoc.Fragments, fragment)
			}
		}
	}

	// Only try to de-dup query fields, since mutations typically have side effects.
	dedup := toDoc.Operations[0].Type == schema.Query
	path := &bytes.Buffer{}
	cache := map[string]schema.Selection{}
	for _, operation := range toDoc.Operations {
		var counter int32 = 0
		operation.Selections = mergeQuerySelections(toDoc, operation.Selections, 'f', &counter, dedup, path, cache)
	}

	var counter int32 = 0
	for i, fragment := range toDoc.Fragments {
		copy := *fragment
		copy.Selections = mergeQuerySelections(toDoc, fragment.Selections, 'F', &counter, dedup, path, map[string]schema.Selection{})
		toDoc.Fragments[i] = &copy
	}
	return toDoc
}

func mergeQuerySelections(doc *schema.QueryDocument, from schema.SelectionList, fieldPrefix rune, counter *int32, dedup bool, path *bytes.Buffer, idx map[string]schema.Selection) schema.SelectionList {
	if from == nil {
		return nil
	}
	result := schema.SelectionList{}

	for _, sel := range from {
		switch original := sel.(type) {
		case *schema.FieldSelection:
			resetPosition := path.Len()
			path.WriteRune('/')
			path.WriteString(original.Name)
			original.Arguments.WriteTo(path)
			original.Directives.WriteTo(path)

			if dedup {
				key := path.String()
				if idx[key] == nil {

					merged := *original
					idx[key] = &merged

					if original.Name == "__typename" {
						merged.Alias = fmt.Sprintf("t")
					} else {
						merged.Alias = fmt.Sprintf("%c%x", fieldPrefix, *counter)
						*counter++
					}
					original.Extension = merged.Alias

					nestedCounter := int32(0)
					merged.Extension = &nestedCounter
					merged.Selections = mergeQuerySelections(doc, merged.Selections, 'f', &nestedCounter, dedup, path, idx)
					result = append(result, &merged)

				} else {
					// Collapse dup field
					merged := idx[key].(*schema.FieldSelection)
					original.Extension = merged.Alias
					nestedCounter := merged.Extension.(*int32)
					selection := mergeQuerySelections(doc, original.Selections, 'f', nestedCounter, dedup, path, idx)
					merged.Selections = append(merged.Selections, selection...)
				}
			} else {
				merged := *original
				if merged.Name == "__typename" {
					merged.Alias = fmt.Sprintf("t")
				} else {
					merged.Alias = fmt.Sprintf("%c%x", fieldPrefix, *counter)
					*counter++
				}
				original.Extension = merged.Alias
				nestedCounter := int32(0)
				merged.Selections = mergeQuerySelections(doc, merged.Selections, 'f', &nestedCounter, dedup, path, idx)
				result = append(result, &merged)
			}
			path.Truncate(resetPosition) // reset it..

		case *schema.InlineFragment:

			resetPosition := path.Len()
			path.WriteRune('/')
			path.WriteString("... on ")
			original.On.WriteTo(path)
			key := path.String()

			if existing, ok := idx[key]; !ok {
				copy := *original
				result = append(result, &copy)
				idx[key] = &copy
				copy.Selections = mergeQuerySelections(doc, copy.Selections, fieldPrefix, counter, dedup, path, idx)
			} else {
				existing := existing.(*schema.InlineFragment)
				existing.Selections = mergeQuerySelections(doc, original.Selections, fieldPrefix, counter, dedup, path, idx)
			}
			path.Truncate(resetPosition) // reset it..

		case *schema.FragmentSpread:

			resetPosition := path.Len()
			path.WriteRune('/')
			path.WriteString("...")
			path.WriteString(original.Name)
			key := path.String()
			if _, ok := idx[key]; !ok {
				result = append(result, original)
				idx[key] = original
			}
			path.Truncate(resetPosition) // reset it..
		}
	}
	return result
}
