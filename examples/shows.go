package main

import (
	"time"

	"github.com/aerogear/graphql-link/examples/shows"
	"github.com/aerogear/graphql-link/internal/gateway"
)

func main() {

	showsEngine := shows.New()
	showsServer, err := gateway.StartServer("localhost", 8082, showsEngine, gateway.SimpleLog)
	if err != nil {
		panic(err)
	}
	defer showsServer.Close()

	for {
		time.Sleep(time.Hour)
	}
}
