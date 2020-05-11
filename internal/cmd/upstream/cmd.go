package new

import (
	"io/ioutil"
	"log"
	"net/url"
	"path/filepath"

	"github.com/chirino/graphql-gw/internal/cmd/root"
	"github.com/chirino/graphql-gw/internal/cmd/serve"
	"github.com/chirino/graphql-gw/internal/gateway"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "add-upstream",
		Short: "Adds new upstream to config.",
		Long:  `Command lets you assemble gateway config by letting you add new upstream gateway`,
		Run:   run,
		Args:  cobra.MinimumNArgs(1),
	}
	UpstreamName = ""
	UpstreamURL  = ""
	ConfigFile   = ""
)

func init() {
	Command.Flags().StringVar(&ConfigFile, "config", "graphql-gw.yaml", "path to the config file to load")
	Command.Flags().StringVar(&UpstreamName, "name", "", "name of the upstream")
	Command.Flags().StringVar(&UpstreamURL, "url", "", "url to the upstream")
	root.Command.AddCommand(Command)
}
func run(cmd *cobra.Command, args []string) {
	vebosityFmt := "%v"
	if !root.Verbose {
		vebosityFmt = "%+v\n"
	}

	file, err := ioutil.ReadFile(ConfigFile)

	if err != nil {
		log.Fatalf("Error reading config file: %s.", err)
		panic(err)
	}

	if len(UpstreamName) > 0 {
		log.Fatalf("Name parameter is missing")
		return
	}

	if len(UpstreamURL) > 0 {
		log.Fatalf("Name parameter is missing")
		return
	}

	url, err := url.Parse(UpstreamURL)
	if err != nil {
		log.Fatalf(vebosityFmt, err)
		panic(err)
	}

	log.Printf(`Contacting upstream %s`, url)

	// TODO refactor loadEndpointSchema

	config := serve.Config{}
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		log.Fatalf(vebosityFmt, "Error parsing yaml file: %s.", err)
		panic(err)
	}

	// TODO define upstream
	config.Upstreams[UpstreamName] = gateway.UpstreamWrapper{}

	for _, typeConfig := range config.Types {
		if typeConfig.Name == "Query" {
			// TODO define actions
			// - name: Query
			// actions:
			//   # mounts all the fields of the root query onto the Query type
			//   - type: mount
			// 	upstream: ` + UpstreamName + `
			// 	query: query {}
			typeConfig.Actions = append(typeConfig.Actions, gateway.ActionWrapper{})
		}
	}

	configYml, err := yaml.Marshal(&config)
	configFile := filepath.Join("./", ConfigFile)
	err = ioutil.WriteFile(configFile, configYml, 0644)

	if err != nil {
		log.Fatalf("%+v", err)
	}

	log.Printf(`Added new upstream to the config`)
	log.Printf(`Review '%s' and then run:`, configFile)
	log.Println()
	log.Println(`    graphql-gw serve`)
	log.Println()
}
