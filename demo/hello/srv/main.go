package main

import (
	"demo/hello/pb"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/tiantour/protector"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	service = flag.String("service", "hello_service", "service name")
	host    = flag.String("host", "http://127.0.0.1", "register etcd address")
	port    = flag.String("port", ":50000", "listening port")
)

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", *port)
	if err != nil {
		panic(err)
	}
	_ = protector.NewRegister().Server(service, host, port)
	log.Printf("starting hello service at %s", *port)
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	s.Serve(lis)
}

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Printf("%v: Receive is %s\n", time.Now(), in.Name)
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}
