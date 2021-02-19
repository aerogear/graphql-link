package main

import (
	"log"
	"time"

	"github.com/aerogear/graphql-link/examples/characters"
	"github.com/aerogear/graphql-link/examples/shows"
	starwars_starship "github.com/aerogear/graphql-link/examples/starwars-starship"
	"github.com/aerogear/graphql-link/examples/starwars_characters"
	"github.com/aerogear/graphql-link/internal/gateway"
)

func main() {
	log.Println("===== Characters =====")
	charactersEngine := characters.New()
	charactersServer, err := gateway.StartServer("0.0.0.0", 8081, charactersEngine, gateway.SimpleLog)
	if err != nil {
		panic(err)
	}
	defer charactersServer.Close()

	log.Println("===== Shows =====")
	showsEngine := shows.New()
	showsServer, err := gateway.StartServer("0.0.0.0", 8082, showsEngine, gateway.SimpleLog)
	if err != nil {
		panic(err)
	}
	defer showsServer.Close()

	log.Println("===== Starwars Characters =====")
	starwars_charactersEngine := starwars_characters.New()
	starwars_charactersServer, err := gateway.StartServer("0.0.0.0", 8083, starwars_charactersEngine, gateway.SimpleLog)
	if err != nil {
		panic(err)
	}
	defer starwars_charactersServer.Close()

	log.Println("===== Starwars StarShip =====")
	starwars_starshipEngine := starwars_starship.New()
	starwars_starshipServer, err := gateway.StartServer("0.0.0.0", 8084, starwars_starshipEngine, gateway.SimpleLog)
	if err != nil {
		panic(err)
	}
	defer starwars_starshipServer.Close()

	for {
		time.Sleep(time.Hour)
	}
}
