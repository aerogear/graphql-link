package main

import (
	"time"

	"github.com/aerogear/graphql-link/examples/starwars_characters"
	"github.com/aerogear/graphql-link/internal/gateway"
)

func main() {

	starwars_charactersEngine := starwars_characters.New()
	starwars_charactersServer, err := gateway.StartServer("localhost", 8083, starwars_charactersEngine, gateway.SimpleLog)
	if err != nil {
		panic(err)
	}
	defer starwars_charactersServer.Close()

	for {
		time.Sleep(time.Hour)
	}
}
