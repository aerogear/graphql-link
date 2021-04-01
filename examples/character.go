package main

import (
	"time"

	"github.com/aerogear/graphql-link/examples/characters"
	"github.com/aerogear/graphql-link/internal/gateway"
)

func main() {

	charactersEngine := characters.New()
	charactersServer, err := gateway.StartServer("localhost", 8081, charactersEngine, gateway.SimpleLog)
	if err != nil {
		panic(err)
	}
	defer charactersServer.Close()

	for {
		time.Sleep(time.Hour)
	}
}
