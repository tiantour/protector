package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/tiantour/protector"
	"github.com/tiantour/protector/example/hello/pb"
	clientv3 "go.etcd.io/etcd/client/v3"
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

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"0.0.0.0:2379"},
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := protector.Register(ctx, client, *srv, *host+*port)
		if err != nil {
			log.Fatal(err)
		}
	}()

	lis, err := net.Listen("tcp", *port)
	if err != nil {
		panic(err)
	}

	log.Printf("starting hello service at %s", *host+*port)
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
