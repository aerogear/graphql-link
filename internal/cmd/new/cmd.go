package new

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/chirino/graphql-gw/internal/cmd/root"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "new",
		Short: "creates a new project with default config",
		Run:   run,
		Args:  cobra.ExactArgs(1),
	}
	ConfigFile = ""
)

func init() {
	root.Command.AddCommand(Command)
}

func run(cmd *cobra.Command, args []string) {
	dir := args[0]
	os.MkdirAll(dir, 0755)

	configFile := filepath.Join(dir, "graphql-gw.yaml")
	err := ioutil.WriteFile(configFile, []byte(`# ------------------------------------------------
# graphql-gw config docs: https://bit.ly/2L5TgyB
# ------------------------------------------------
listen: localhost:8080

##
## Example config to configure an upstream server and expose it's 
## Query type via the gateway's Query type.
##
# upstreams:
#   anilist:
#     url: https://graphql.anilist.co/
#     prefix: Ani
# 
# types:
#   - name: Query
#     actions:
#       - type: mount
#         upstream: anilist
#         query: query {}
# 

`), 0644)

	if err != nil {
		log.Fatalf("%+v", err)
	}

	log.Printf(`Project created in the '%s' directory.`, dir)
	log.Printf(`Edit '%s' and then run:`, configFile)
	log.Println()
	log.Println(`    cd`, dir)
	log.Println(`    graphql-gw serve`)
	log.Println()
}
