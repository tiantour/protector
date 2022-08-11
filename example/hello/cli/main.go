package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"strconv"

	"github.com/tiantour/protector"
	"github.com/tiantour/protector/example/hello/pb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

var (
	srv = flag.String("service", "hello_service", "service name")
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
	b := protector.NewBuilder(client)
	resolver.Register(b)

	target := fmt.Sprintf("%s://%s", protector.Scheme, *srv)
	conn, err := grpc.Dial(
		target,
		grpc.WithBlock(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	cli := pb.NewGreeterClient(conn)

	ticker := time.NewTicker(1 * time.Second)
	for t := range ticker.C {
		b.DebugStore()

		resp, err := cli.SayHello(context.Background(), &pb.HelloRequest{
			Name: "world " + strconv.Itoa(t.Second()),
		})
		if err == nil {
			fmt.Printf("%v: Reply is %s\n", t, resp.Message)
		} else {
			fmt.Println(err)
		}
	}
}
