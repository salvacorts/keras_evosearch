package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"server/ea"
	pb "server/protobuf/api"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/MaxHalford/eaopt"

	stdlog "log"

	log "github.com/sirupsen/logrus"
)

type ApiServer struct {
	pb.UnimplementedAPIServer
}

func (s *ApiServer) GetModelParams(ctx context.Context, in *pb.Empty) (*pb.ModelParameters, error) {
	log.Debug("Sending new model")

	for _, c := range ea.Models_chan_to_evaluate {
		if len(c) == 0 {
			continue
		}

		model := <-c

		log.Debug("Sending model (%s) to evaluate\n", model.ModelId)
		return model, nil
	}

	return nil, status.Errorf(1, "No models to evaluate")
}

func (s *ApiServer) ReturnModel(ctx context.Context, results *pb.ModelResults) (*pb.Empty, error) {
	log.Debug("Model (%s) returned. Recall: %f\n", results.ModelId, results.Recall)

	ea.Models_chan_evaluated[results.ModelId] <- results

	return &pb.Empty{}, nil
}

func main() {
	listen_addr := fmt.Sprintf("0.0.0.0:%d", 10000)
	lis, err := net.Listen("tcp", listen_addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Info("Listening at " + listen_addr)

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterAPIServer(grpcServer, &ApiServer{})
	go grpcServer.Serve(lis)

	ga, err := eaopt.NewDefaultGAConfig().NewGA()
	if err != nil {
		fmt.Println(err)
		return
	}

	ga.NGenerations = 30
	ga.PopSize = 50
	ga.ParallelEval = true
	ga.Logger = stdlog.New(os.Stdout, "", stdlog.Ldate|stdlog.Ltime)
	ga.EarlyStop = func(ga *eaopt.GA) bool {
		return ga.HallOfFame[0].Fitness == 0
	}
	ga.Callback = func(ga *eaopt.GA) {
		log.Info("Best fitness at generation %d: %f",
			ga.Generations, ga.HallOfFame[0].Fitness)
	}
	best := ga.Minimize(ea.MakeModel)

	log.Info("Best model found:\n%x", best)
}
