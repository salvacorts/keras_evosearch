package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "server/protobuf/api"

	grpc "google.golang.org/grpc"
)

type ApiServer struct {
	pb.UnimplementedAPIServer

	ModelId uint64
}

func (s *ApiServer) GetModelParams(ctx context.Context, in *pb.Empty) (*pb.ModelParameters, error) {
	fmt.Println("Sending new model")

	return &pb.ModelParameters{
		ModelId: s.ModelId,

		LearningRate:   0.01,
		Optimizer:      pb.Optimizers_Adam,
		ActivationFunc: "relu",

		Layers: []*pb.Layer{
			{
				NumNeurons: 32,
			},
			{
				NumNeurons: 32,
			},
		},
	}, nil
}

func (s *ApiServer) ReturnModel(ctx context.Context, results *pb.ModelResults) (*pb.Empty, error) {
	fmt.Printf("Model (%d) returned. Recall: %f\n", results.ModelId, results.Recall)

	s.ModelId += 1
	return &pb.Empty{}, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 10000))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterAPIServer(grpcServer, &ApiServer{})
	grpcServer.Serve(lis)
}
