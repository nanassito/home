package main

import (
	"fmt"
	"log"
	"net"

	pb "github.com/nanassito/home/proto/go/switches"
	switches "github.com/nanassito/home/switches/packages"
	grpc "google.golang.org/grpc"
)

func main() {
	addr := "localhost:7001"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	switchesServer := switches.Server{}
	pb.RegisterSwitchSvcServer(grpcServer, switchesServer)
	fmt.Println("Server started on " + addr)
	grpcServer.Serve(lis)
}
