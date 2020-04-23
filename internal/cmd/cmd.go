package cmd

import (
	_ "github.com/chirino/graphql-gw/internal/cmd/new"
	"github.com/chirino/graphql-gw/internal/cmd/root"
	_ "github.com/chirino/graphql-gw/internal/cmd/serve"
)

func Main() {
	root.Main()
}
