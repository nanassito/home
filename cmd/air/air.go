package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/nanassito/home/pkg/air"
	"github.com/nanassito/home/pkg/air_proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var logger = log.New(os.Stderr, "", log.Lshortfile)

func main() {
	// Serve Grpc API
	grpcAddr := ":7006"
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	airServer := air.NewServer()
	air_proto.RegisterAirSvcServer(grpcServer, airServer)
	go grpcServer.Serve(lis)

	// Serve Prometheus metrics
	prometheus.MustRegister(&air.PromCollector{Server: airServer})
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":7005", nil)

	// Serve Http Proxy for the API
	conn, err := grpc.DialContext(
		context.Background(),
		grpcAddr,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Fatalln("Failed to dial grpc server:", err)
	}
	gwmux := runtime.NewServeMux()
	err = air_proto.RegisterAirSvcHandler(context.Background(), gwmux, conn)
	if err != nil {
		logger.Fatalln("Failed to register gateway:", err)
	}
	gwServer := &http.Server{Addr: ":7007", Handler: gwmux}
	logger.Println("Info| Air server started")
	logger.Fatalln(gwServer.ListenAndServe())
}
