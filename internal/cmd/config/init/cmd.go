package init

import (
	"io/ioutil"
	"os"

	"github.com/aerogear/graphql-link/internal/cmd/config"
	"github.com/aerogear/graphql-link/internal/cmd/root"
	"github.com/aerogear/graphql-link/internal/gateway"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "init",
		Short: "creates the gateway default configuration file",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// to override the PersistentPreRun in the config.Command
		},
		Run: run,
	}
)

func init() {
	config.Command.AddCommand(Command)
}

func run(_ *cobra.Command, _ []string) {
	log := gateway.SimpleLog

	if _, err := os.Stat(config.File); err == nil {
		log.Fatalf("error: file exists %s", config.File)
	}

	err := ioutil.WriteFile(config.File, []byte(`# ------------------------------------------------
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
types:
  - name: Query
    actions:
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
  - name: Mutation
    actions:
#      # mounts all the fields of the root anilist mutation
#      - type: mount
#        upstream: anilist
#        query: mutation {}
  - name: Subscription
    actions:
`), 0644)

	if err != nil {
		log.Fatalf(root.Verbosity, err)
	}

	log.Println()
	log.Println(`Created: `, config.File)
	log.Println()
	log.Println(`Start the gateway by running:`)
	log.Println()
	log.Println(`    graphql-gw serve`)
	log.Println()
}
