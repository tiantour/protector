package protector

import (
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
)

// Resolver resolver
type Resolver struct {
	Target string
}

// NewResolver new resolver
func NewResolver() *Resolver {
	return &Resolver{}
}

// Resolve Resolve
func (r *Resolver) Resolve(target string) (naming.Watcher, error) {
	cli := NewProtector().Client()
	return &Watcher{
		Client: cli,
		Target: r.Target,
	}, nil
}

// Client resolver client
func (r *Resolver) Client(srv string) (*grpc.ClientConn, error) {
	r.Target = fmt.Sprintf("/%s/%s/", Prefix, srv)
	b := grpc.RoundRobin(r)
	return grpc.DialContext(context.Background(), Endpoints[0], grpc.WithInsecure(), grpc.WithBalancer(b), grpc.WithBlock())
}
