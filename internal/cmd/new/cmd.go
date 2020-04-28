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
	err := ioutil.WriteFile(configFile, []byte(`#
# Configure the host and port the service will listen on
listen: localhost:8080

#
# Configure the GraphQL endpoints you will be composing here along with their schema.
endpoints:
  anilist:
    url: https://graphql.anilist.co/
    prefix: Ani

types:
  - name: Query
    fields:
      # imports all the fields into Query type.
      - endpoint: anilist
        query: query {}

      # Adds only a test_field
      - endpoint: anilist
        query: query {}
        name: ani_query

      # Adds a animeCharacters($page:Int, $perPage:Int, $search:String) field
      - endpoint: anilist
        query: |
          query ($page:Int, $perPage:Int, $search:String) {
            Page(page:$page, perPage:$perPage) {
              characters(search:$search)
            }
          }
        name: animeCharacters

  - name: Mutation
    fields:
      - endpoint: anilist
        query: mutation {}
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
