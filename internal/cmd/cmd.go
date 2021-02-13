package cmd

import (
	_ "github.com/aerogear/graphql-link/internal/cmd/completion"
	_ "github.com/aerogear/graphql-link/internal/cmd/config/add/upstream"
	_ "github.com/aerogear/graphql-link/internal/cmd/config/init"
	_ "github.com/aerogear/graphql-link/internal/cmd/config/link"
	_ "github.com/aerogear/graphql-link/internal/cmd/config/mount"
	"github.com/aerogear/graphql-link/internal/cmd/root"
	_ "github.com/aerogear/graphql-link/internal/cmd/serve"
	"github.com/aerogear/graphql-link/internal/cmd/version"
)

type VersionConfig = version.VersionConfig

func Main(versionConfig VersionConfig) {
	version.Config = versionConfig
	root.Main()
}
