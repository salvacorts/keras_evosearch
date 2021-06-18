package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

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

	log.Warn("No models to evaluate")
	return nil, status.Errorf(codes.Canceled, "No models to evaluate")
}

func (s *ApiServer) ReturnModel(ctx context.Context, results *pb.ModelResults) (*pb.Empty, error) {
	log.Debugf("Model (%s) returned. Recall: %f", results.ModelId, results.Recall)

	ea.Models_chan_evaluated[results.ModelId] <- results

	return &pb.Empty{}, nil
}

func SetupLogger() {

}

func main() {
	listen_addr := flag.String("listen", "0.0.0.0:10000", "Address to listen at")
	log_level := flag.Int("verbosity", 4, "Verbosity level (Default info)")
	flag.Parse()

	// Setup logger
	now := time.Now()
	timespampFormat := "Jan__2_15_04_05"
	fileName := fmt.Sprintf("logs/%s/generations.log", now.Format(timespampFormat))
	_ = os.MkdirAll(filepath.Dir(fileName), os.ModePerm)
	var file, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Panic("Failed to open log file")
	}
	defer file.Close()

	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.Level(*log_level))

	// Setup gRPC
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
	ga.ParallelEval = true // true
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
	err = ga.Minimize(ea.MakeModel)
	if err != nil {
		log.Panicf("GA failed. %e", err)
	}

	best := ga.HallOfFame[0].Genome.(*ea.ModelGenome)
	log.Infof("Best model found: %s", best.String())
}
