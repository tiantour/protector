package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"strconv"

	"github.com/tiantour/protector"
	"github.com/tiantour/protector/example/hello/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

var (
	srv = flag.String("srv", "hello_service", "service name")
)

func init() {
	b := protector.NewResolver()
	resolver.Register(b)
}

func main() {
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	target := fmt.Sprintf("%s:///%s", protector.Scheme, *srv)
	conn, err := grpc.DialContext(
		ctx,
		target,
		grpc.WithBlock(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal("resolver etcd err", err)
	}
	defer conn.Close()

	cli := pb.NewGreeterClient(conn)

	ticker := time.NewTicker(1 * time.Second)
	for t := range ticker.C {

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
