package main

import (
	"fmt"
	"log"
	"net"
	_ "net/http/pprof"
	"os"

	"payment/config"
	pb "payment/pb"
	"payment/server"

	"google.golang.org/grpc"
)

var (
	banner = `　　　　 ____ ____ ____ ____ ____ ____ ____ ____ ____ ____
	||P |||a |||y |||m |||e |||n |||t |||A |||P |||I ||
	||__|||__|||__|||__|||__|||__|||__|||__|||__|||__||
	|/__\|/__\|/__\|/__\|/__\|/__\|/__\|/__\|/__\|/__\|`
)

type PaymentService struct{}

func main() {
	fmt.Println(banner)

	httpPort := os.Getenv("PAYMENT_HTTP_PORT")
	if httpPort == "" {
		httpPort = "0.0.0.0:5000"
	}
	grpcPort := os.Getenv("PAYMENT_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "0.0.0.0:5001"
	}

	//setup config
	c := config.Config{
		HttpPort: httpPort,
		GrpcPort: grpcPort,
	}
	log.Printf("HTTP Port%s, gRPC Port%s\n", c.HttpPort, c.GrpcPort)

	//setup grpc server
	lis, err := net.Listen("tcp", c.GrpcPort)
	if err != nil {
		log.Fatalf("listen error: %s\n", err)
	}
	g := grpc.NewServer()

	s, err := server.NewNetworkServer()
	if err != nil {
		log.Fatalf("failed to create new server:%s", err)
	}

	pb.RegisterPaymentServiceServer(g, s)
	done := make(chan struct{})
	go func() {
		err = g.Serve(lis)
		if err != nil {
			log.Fatal(err)
		}
		done <- struct{}{}
	}()

	go func() {
		if err := server.StartGRPCGateway(c); err != nil {
			log.Fatal(err)
		}
		done <- struct{}{}
	}()
	<-done // waiting finish goroutine

	log.Fatal("Program exit")
}
