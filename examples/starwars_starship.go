package main

import (
	"time"

	starwars_starship "github.com/aerogear/graphql-link/examples/starwars-starship"
	"github.com/aerogear/graphql-link/internal/gateway"
)

func main() {

	starwars_starshipEngine := starwars_starship.New()
	starwars_starshipServer, err := gateway.StartServer("localhost", 8084, starwars_starshipEngine, gateway.SimpleLog)
	if err != nil {
		panic(err)
	}
	defer starwars_starshipServer.Close()

	for {
		time.Sleep(time.Hour)
	}
}
