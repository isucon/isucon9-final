package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	_ "net/http/pprof"

	"github.com/chibiegg/isucon9-final/blackbox/payment/config"
	pb "github.com/chibiegg/isucon9-final/blackbox/payment/pb"
	"github.com/chibiegg/isucon9-final/blackbox/payment/server"
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

	//setup config
	configFile := flag.String("config-file", "/etc/paymentapi/config.yml", "config file path")
	flag.Parse()
	c, err := config.LoadFile(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("HTTP Port%s, gRPC Port%s\n", c.HttpPort, c.GrpcPort)

	//setup grpc server
	lis, err := net.Listen("tcp", c.GrpcPort)
	if err != nil {
		log.Fatalf("listen error: \n", err)
	}
	g := grpc.NewServer()

	s, err := server.NewNetworkServer(c)
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
