//go:generate go run .

package main

import (
	"os"

	"github.com/chirino/hawtgo/sh"
)

func main() {

	sh.New().
		Dir("..").
		CommandLog(os.Stdout).
		CommandLogPrefix(`go >`).
		LineArgs(`go`, `install`,
			//`github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway`,
			//`github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger`,
			`github.com/golang/protobuf/protoc-gen-go`).
		MustZeroExit()

	sh.New().
		Dir("..").
		CommandLog(os.Stdout).
		CommandLogPrefix("protoc > ").
		Line(`protoc -I. --go_out=plugins=grpc,paths=source_relative:./proto service.proto`).
		MustZeroExit()

	//sh.New().
	//	Dir("..").
	//	CommandLog(os.Stdout).
	//	CommandLogPrefix("protoc > ").
	//	Line(`protoc -I. --grpc-gateway_out=logtostderr=true,grpc_api_configuration=service.yaml,paths=source_relative:./proto service.proto`).
	//	MustZeroExit()
	//
	//sh.New().
	//	Dir("..").
	//	CommandLog(os.Stdout).
	//	CommandLogPrefix("protoc > ").
	//	Line(`protoc -I. --swagger_out=logtostderr=true,grpc_api_configuration=service.yaml:./proto service.proto`).
	//	MustZeroExit()

}
