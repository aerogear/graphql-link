package new

import (
	"io/ioutil"
	"log"
	"net/url"
	"path/filepath"

	"github.com/chirino/graphql-gw/internal/cmd/root"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "setup",
		Short: "creates a new project and helps you to assemble config using.",
		Long: `Command lets you assemble gateway config by letting you specify list of the servers gateway will connect with. 
    By default gateway will expose all Mutations, Queries and Subscriptions from your server`,
		Run:  run,
		Args: cobra.MinimumNArgs(1),
	}
	ConfigFile = ""
)

func init() {
	root.Command.AddCommand(Command)
}

func run(cmd *cobra.Command, args []string) {
	upstreams := ""
	types := ""
	for _, s := range args {
		name, err := url.Parse(s)
		if err != nil {
			log.Fatal(err)
		}
		// Replace with templates
		upstreams +=
			"\n   " + name.Hostname() + `:
      url: ` + s + "\n"

		types +=
			`
      - name: Query
        actions:
          # mounts all the fields of the root query onto the Query type
          - type: mount
            upstream: ` + name.Hostname() + `
            query: query {}

      - name: Mutation
        actions:
          # mounts all the fields of the root query onto the Mutation type
          - type: mount
            upstream: ` + name.Hostname() + `
            query: mutation {}

      - name: Subscription
        actions:
          # mounts all the fields of the root query onto the Subscription type
          - type: mount
            upstream: ` + name.Hostname() + `
            query: subscription {}
        
        `
	}

	configFile := filepath.Join("./", "graphql-gw.yaml")
	err := ioutil.WriteFile(configFile, []byte(`#
# Configure the host and port the service will listen on
listen: localhost:8080

#
# Configure the GraphQL upstream servers you will be accessing
upstreams: `+upstreams+`
types:`+types+"\n"), 0644)

	if err != nil {
		log.Fatalf("%+v", err)
	}

	log.Printf(`Project created in the current directory.`)
	log.Printf(`Review '%s' and then run:`, configFile)
	log.Println()
	log.Println(`    graphql-gw serve`)
	log.Println()
}
