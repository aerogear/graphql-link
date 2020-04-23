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
		Short: "creates a graphql-gw project with default config",
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
listen: 0.0.0.0:8080

#
# Configure the GraphQL endpoints you will be composing here along with their schema.
endpoints:
  # # Example Endpoint
  # ep1:
  #   url: http://localhost:8081/

query:
  # # Adding an example query field that forwards to an endpoint.
  # hi:
  #   endpoint: ep1 # a reference to an endpoint configured above
  #   description: provided by the hello service
  #   query: |
  #     query($tok: String!, $firstName:String!) {
  #       login(token:$tok) {
  #         hello(name:$firstName)
  #       }
  #     }

`), 0644)

	if err != nil {
		log.Fatalf("%+v", err)
	}

	log.Printf(`Project created in the '%s' directory.`, dir)
	log.Printf(`Edit '%s' and then run:`, configFile)
	log.Println()
	log.Println(`    cd`, dir)
	log.Println(`    grapql-gw serve`)
	log.Println()
}
