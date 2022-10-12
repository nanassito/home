package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/nanassito/home/pkg/air"
	"github.com/nanassito/home/pkg/air_proto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Serve Prometheus metrics
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":7005", nil)

	// Serve Grpc API
	grpcAddr := ":7006"
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	airServer := air.New()
	air_proto.RegisterAirSvcServer(grpcServer, airServer)
	go grpcServer.Serve(lis)

	// Serve Http Proxy for the API
	conn, err := grpc.DialContext(
		context.Background(),
		grpcAddr,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial grpc server:", err)
	}
	gwmux := runtime.NewServeMux()
	err = air_proto.RegisterAirSvcHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}
	gwServer := &http.Server{Addr: ":7007", Handler: gwmux}
	fmt.Println("Air server started")
	log.Fatalln(gwServer.ListenAndServe())
}
