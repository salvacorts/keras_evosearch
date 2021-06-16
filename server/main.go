package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"

	"server/ea"
	pb "server/protobuf/api"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/MaxHalford/eaopt"

	stdlog "log"

	log "github.com/sirupsen/logrus"
)

type ApiServer struct {
	pb.UnimplementedAPIServer
}

func (s *ApiServer) GetModelParams(ctx context.Context, in *pb.Empty) (*pb.ModelParameters, error) {
	for _, c := range ea.Models_chan_to_evaluate {
		if len(c) == 0 {
			continue
		}

		model := <-c

		log.Debugf("Sending model (%s) to evaluate", model.ModelId)
		return model, nil
	}

	// log.Errorf("No models to evaluate, Models_chan_to_evaluate size=%d",
	// 	len(ea.Models_chan_to_evaluate))

	// for id, c := range ea.Models_chan_to_evaluate {
	// 	fmt.Printf("Chan evaluate (%s) - len=%d cap=%d\n", id, len(c), cap(c))
	// }
	// for id, c := range ea.Models_chan_evaluated {
	// 	fmt.Printf("Chan evaluated (%s) - len=%d cap=%d\n", id, len(c), cap(c))
	// }

	return nil, status.Errorf(codes.Canceled, "No models to evaluate")
}

func (s *ApiServer) ReturnModel(ctx context.Context, results *pb.ModelResults) (*pb.Empty, error) {
	log.Debugf("Model (%s) returned. Recall: %f", results.ModelId, results.Recall)

	ea.Models_chan_evaluated[results.ModelId] <- results

	return &pb.Empty{}, nil
}

func main() {
	listen_addr := flag.String("listen", "0.0.0.0:10000", "Address to listen at")
	log_level := flag.Int("verbosity", 4, "Verbosity level (Default info)")
	flag.Parse()

	log.SetLevel(log.Level(*log_level))

	lis, err := net.Listen("tcp", *listen_addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Info("Listening at " + *listen_addr)

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterAPIServer(grpcServer, &ApiServer{})
	go grpcServer.Serve(lis)
	defer grpcServer.GracefulStop()

	ga, err := eaopt.NewDefaultGAConfig().NewGA()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Configure the GA
	ga.NGenerations = 30
	ga.PopSize = 50
	ga.ParallelEval = true
	ga.Model = eaopt.ModGenerational{
		Selector: eaopt.SelTournament{
			NContestants: 2,
		},

		// Since we apply several Mutation and Cross funcions
		// We decide whether to perform them or not within the functions
		MutRate:   1,
		CrossRate: 1,
	}
	ga.Logger = stdlog.New(os.Stdout, "", stdlog.Ldate|stdlog.Ltime)
	ga.EarlyStop = func(ga *eaopt.GA) bool {
		return ga.HallOfFame[0].Fitness == 0
	}
	ga.Callback = func(ga *eaopt.GA) {
		log.Infof("Best fitness at generation %d: %f",
			ga.Generations, ga.HallOfFame[0].Fitness)
	}

	// Start the GA
	best := ga.Minimize(ea.MakeModel)
	log.Info("Best model found:\n%x", best)
}
