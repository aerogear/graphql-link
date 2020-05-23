package gateway

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql/schema"
	"github.com/pkg/errors"
)

func loadEndpointSchema(config Config, upstream *upstreamServer) (*schema.Schema, error) {

	schemaText := upstream.info.Schema
	if strings.TrimSpace(schemaText) != "" {
		config.Log.Printf("using static schema for upstream %s: %s", upstream.id, upstream.info.URL)
		return Parse(schemaText)
	}

	upstreamSchemaFile := filepath.Join(config.ConfigDirectory, "upstreams", upstream.id+".graphql")
	upstreamSchemaFileExists := false
	if stat, err := os.Stat(upstreamSchemaFile); err == nil && !stat.IsDir() {
		upstreamSchemaFileExists = true
	}

	if upstreamSchemaFileExists {
		data, err := ioutil.ReadFile(upstreamSchemaFile)
		if err == nil {
			r, err := Parse(string(data))
			if err == nil {
				config.Log.Printf("loaded previously stored schema: %s", upstreamSchemaFile)
				return r, err
			}
		}
	}

	if !config.DisableSchemaDownloads {
		config.Log.Printf("downloading schema for upstream %s: %s", upstream.id, upstream.info.URL)
		s, err := downloadSchema(config, upstream)
		if err != nil {
			if upstreamSchemaFileExists {
				config.Log.Printf("download failed (will load cached schema version): %v", err)
			} else {
				return nil, errors.Wrap(err, "download failed")
			}
		} else {
			return s, nil
		}
	}

	if upstreamSchemaFileExists {
		config.Log.Printf("loading previously stored schema: %s", upstreamSchemaFile)
		// This could be a transient failure... see if we have previously save it's schema.
		data, err := ioutil.ReadFile(upstreamSchemaFile)
		if err != nil {
			return nil, err
		}
		return Parse(string(data))
	}

	return nil, errors.Errorf("no schema defined for upstream %s: %s", upstream.id, upstream.info.URL)
}

func downloadSchema(config Config, upstream *upstreamServer) (*schema.Schema, error) {

	s, err := graphql.GetSchema(upstream.client)

	// We may need to store it if it succeeded.
	if err == nil && config.EnabledSchemaStorage {
		upstreamSchemaFile := filepath.Join(config.ConfigDirectory, "upstreams", upstream.id+".graphql")
		err := ioutil.WriteFile(upstreamSchemaFile, []byte(s.String()), 0644)
		if err != nil {
			return nil, errors.Wrap(err, "could not update schema")
		}
	}
	return s, err

}

func Parse(schemaText string) (*schema.Schema, error) {
	s := schema.New()
	err := s.Parse(schemaText)
	if err != nil {
		return nil, err
	}
	return s, nil
}
