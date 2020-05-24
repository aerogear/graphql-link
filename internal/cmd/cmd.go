package cmd

import (
	_ "github.com/chirino/graphql-gw/internal/cmd/new"
	"github.com/chirino/graphql-gw/internal/cmd/root"
	_ "github.com/chirino/graphql-gw/internal/cmd/serve"
	_ "github.com/chirino/graphql-gw/internal/cmd/upstream"
	"github.com/chirino/graphql-gw/internal/cmd/version"
)

type VersionConfig = version.VersionConfig

func Main(versionConfig VersionConfig) {
	version.Config = versionConfig
	root.Main()
}
