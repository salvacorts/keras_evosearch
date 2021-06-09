package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "server/gen/api"

	grpc "google.golang.org/grpc"
)

type ApiServer struct {
	pb.UnimplementedAPIServer
}

func (ApiServer) SayHello(ctx context.Context, in *pb.Word) (*pb.Word, error) {
	out := fmt.Sprintf("%s 1234", in.Word)
	return &pb.Word{Word: out}, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 10000))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterAPIServer(grpcServer, ApiServer{})
	grpcServer.Serve(lis)
}
