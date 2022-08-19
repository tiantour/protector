package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/tiantour/protector"
	"github.com/tiantour/protector/example/hello/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	srv  = flag.String("srv", "hello_service", "service name")
	port = flag.String("port", ":5000", "listening port")
	host = flag.String("host", "localhost", "register etcd address")
)

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", *port)
	if err != nil {
		panic(err)
	}

	go func() {
		err := protector.Register(*srv, *host, *port)
		if err != nil {
			log.Fatal(err)
		}
	}()

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
