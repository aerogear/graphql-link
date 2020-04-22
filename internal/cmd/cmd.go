package cmd

import (
	_ "github.com/chirino/graphql-gw/internal/cmd/serve"
	_ "github.com/chirino/graphql-gw/internal/cmd/new"
	"github.com/chirino/graphql-gw/internal/cmd/root"
)

func Main() {
	root.Main()
}
