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
## Some example configuration.... 
##
#upstreams:
#  anilist:
#    url: https://graphql.anilist.co
#    prefix: Ani
#  pokemon:
#    url: https://graphql-pokemon.now.sh
#    prefix: Pokemon
#    
#types:
#  - name: Query
#    actions:
#      # mounts the root anilist query to the anime field
#      - type: mount
#        field: anime
#        upstream: anilist
#        query: query {}
#
#      # mounts the root pokemon query to the pokemon field
#      - type: mount
#        field: pokemon
#        upstream: pokemon
#        query: query {}
#
#  - name: AniCharacter
#    actions:
#      # mounts the root anilist query to the anime field
#      - type: link
#        field: pokemon
#        vars:
#          $fullname: name { full }
#        upstream: pokemon
#        query: query { pokemon(name:$fullname) }
##
## The above link lets you do queries that access data from both the anilist and pokemon services. Example:
##
##     query {
##      anime {
##        Character(search: "Pikachu") {
##          description
##          image {
##            medium
##          }
##          pokemon {
##            attacks {
##              special {
##                name
##                type
##                damage
##              }
##            }
##          }
##        }
##      }
##     }
#
#  - name: Mutation
#    actions:
#      # mounts all the fields of the root anilist mutation
#      - type: mount
#        upstream: anilist
#        query: mutation {}

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
