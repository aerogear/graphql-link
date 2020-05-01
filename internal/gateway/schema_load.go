package gateway

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql/schema"
	"github.com/pkg/errors"
)

func loadEndpointSchema(config Config, eid string, upstream *upstreamServer) (*schema.Schema, error) {

	schemaText := upstream.info.Schema
	if strings.TrimSpace(schemaText) != "" {
		log.Printf("using static schema for upstream %s: %s", eid, upstream.info.URL)
		return Parse(schemaText)
	}

	upstreamSchemaFile := filepath.Join(config.ConfigDirectory, "upstreams", eid+".graphql")
	upstreamSchemaFileExists := false
	if stat, err := os.Stat(upstreamSchemaFile); err == nil && !stat.IsDir() {
		upstreamSchemaFileExists = true
	}

	if !config.DisableSchemaDownloads {
		log.Printf("downloading schema for upstream %s: %s", eid, upstream.info.URL)
		s, err := graphql.GetSchema(upstream.client)

		if err != nil {
			if upstreamSchemaFileExists {
				log.Printf("download failed (will load cached schema version): %v", err)
			} else {
				return nil, errors.Wrap(err, "download failed")
			}
		}

		// We may need to store it if it succeeded.
		if err == nil && config.EnabledSchemaStorage {
			err := ioutil.WriteFile(upstreamSchemaFile, []byte(s.String()), 0644)
			if err != nil {
				return nil, errors.Wrap(err, "could not update schema")
			}
		}

		return s, nil
	}

	if upstreamSchemaFileExists {
		log.Printf("loading previously stored schema: %s", upstreamSchemaFile)
		// This could be a transient failure... see if we have previously save it's schema.
		data, err := ioutil.ReadFile(upstreamSchemaFile)
		if err != nil {
			return nil, err
		}
		return Parse(string(data))
	}

	return nil, errors.Errorf("no schema defined for upstream %s: %s", eid, upstream.info.URL)
}

func Parse(schemaText string) (*schema.Schema, error) {
	s := schema.New()
	err := s.Parse(schemaText)
	if err != nil {
		return nil, err
	}
	return s, nil
}
