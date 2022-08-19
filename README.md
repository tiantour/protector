# protector

grpc service register & resolver with etcd3

# How to use

1. config

    ```golang
    // default scheme
    protector.Scheme    = "etcd"

    // default endpoints
	protector.Endpoints = []string{"0.0.0.0:2379"}

    // default ttl
    protector.TTL = 5
    ```

2. server

    ```golang
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
    ```

3. client

    ```golang
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

        // tips, there must be :///, if not
        // it can't make a distinction between different service
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
    ```