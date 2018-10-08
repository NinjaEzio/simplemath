package main

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	pb "simplemath/api"
	"simplemath/server/rpcimpl"
	"strings"
)

const (
	port = ":50051"
)

// authenticateClient check the client credentials
func authenticateClient(ctx context.Context) (string, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		clientUsername := strings.Join(md["username"], "")
		clientPassword := strings.Join(md["password"], "")
		if clientUsername != "valineliu" {
			return "", fmt.Errorf("unknown user %s", clientUsername)
		}
		if clientPassword != "root" {
			return "", fmt.Errorf("wrong password %s", clientPassword)
		}
		log.Printf("authenticated client: %s", clientUsername)
		return "9527", nil
	}
	return "", fmt.Errorf("missing credentials")
}

// unaryInterceptor calls authenticateClient with current context
func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	clientID, err := authenticateClient(ctx)
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, "clientID", clientID)
	return handler(ctx, req)
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	creds, err := credentials.NewServerTLSFromFile("../cert/server.crt", "../cert/server.key")
	if err != nil {
		log.Fatalf("could not load TLS keys: %s", err)
	}
	// Create an array of gRPC options with the credentials
	opts := []grpc.ServerOption{grpc.Creds(creds), grpc.UnaryInterceptor(unaryInterceptor)}
	s := grpc.NewServer(opts...)
	pb.RegisterSimpleMathServer(s, &rpcimpl.SimpleMathServer{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
