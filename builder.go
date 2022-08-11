package protector

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

const (
	Scheme      = "etcd"
	defaultFreq = time.Minute * 30
)

type Builder struct {
	client *clientv3.Client
	store  map[string]map[string]struct{}
}

func NewBuilder(client *clientv3.Client) *Builder {
	return &Builder{
		client: client,
		store:  make(map[string]map[string]struct{}),
	}
}

func (b *Builder) DebugStore() {
	fmt.Printf("store %+v\n", b.store)
}

// Build creates a new resolver for the given target.
//
// gRPC dial calls Build synchronously, and fails if the returned error is
// not nil.
func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	b.store[target.URL.Path] = make(map[string]struct{})

	r := &Resolver{
		client: b.client,
		target: target,
		cc:     cc,
		store:  b.store[target.URL.Path],
		stopCh: make(chan struct{}, 1),
		rn:     make(chan struct{}, 1),
		t:      time.NewTicker(defaultFreq),
	}

	go r.watch(context.Background())
	r.ResolveNow(resolver.ResolveNowOptions{})

	return r, nil
}

// Scheme returns the scheme supported by this resolver.
// Scheme is defined at https://github.com/grpc/grpc/blob/master/doc/naming.md.
func (b *Builder) Scheme() string {
	return Scheme
}
