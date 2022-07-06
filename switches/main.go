package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	pb "github.com/nanassito/home/proto/switches"
	switches "github.com/nanassito/home/switches/packages"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	grpc "google.golang.org/grpc"
)

func main() {
	addr := "localhost:7001"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":7002", nil)

	grpcServer := grpc.NewServer()
	switchesServer := switches.New()
	pb.RegisterSwitchSvcServer(grpcServer, switchesServer)
	fmt.Println("Switch server started on " + addr)
	grpcServer.Serve(lis)
}
