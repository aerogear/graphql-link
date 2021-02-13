package main

import (
	"log"
	"time"

	"github.com/aerogear/graphql-link/internal/gateway"
	"github.com/aerogear/graphql-link/internal/gateway/examples/characters"
	"github.com/aerogear/graphql-link/internal/gateway/examples/shows"
	"github.com/aerogear/graphql-link/internal/gateway/examples/starwars_characters"
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

	log.Println("===== Starwars =====")
	starwars_charactersEngine := starwars_characters.New()
	starwars_charactersServer, err := gateway.StartServer("0.0.0.0", 8083, starwars_charactersEngine, gateway.SimpleLog)
	if err != nil {
		panic(err)
	}
	defer starwars_charactersServer.Close()

	for {
		time.Sleep(time.Hour)
	}
}
