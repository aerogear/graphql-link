package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"

	gw "github.com/chirino/graphql-gw/internal/gateway/policyagent/proto"
)

var (
	tls      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("cert_file", "", "The TLS cert file")
	keyFile  = flag.String("key_file", "", "The TLS key file")
	grpcPort = flag.Int("grpc-port", 10000, "The server port")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var opts []grpc.ServerOption
	if *tls {
		if *certFile == "" {
			*certFile = testdata.Path("server1.pem")
		}
		if *keyFile == "" {
			*keyFile = testdata.Path("server1.key")
		}
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(opts...)
	gw.RegisterPolicyAgentServer(grpcServer, &server{})

	log.Println("GRPC service available at:", lis.Addr())
	return grpcServer.Serve(lis)

}

////////////////////////////////////////////////////////////////////////////////////////
// server Implements the GRPC PolicyAgentServer service interface:
////////////////////////////////////////////////////////////////////////////////////////
type server struct {
}

func (s *server) Check(ctx context.Context, req *gw.CheckRequest) (*gw.CheckResponse, error) {
	resp := gw.CheckResponse{}
	for i, f := range req.Graphql.GetFields() {
		if (i % 2) == 1 {
			resp.Fields = append(resp.Fields, &gw.GraphQLFieldResponse{
				Path:  f.Path,
				Error: "You are not allowed to access odd fields",
			})
		}
	}
	return &resp, nil
}
