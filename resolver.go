package protector

import (
	"context"
	"path/filepath"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/pkg/errors"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

type Resolver struct {
	client  *clientv3.Client              // client
	target  resolver.Target               // target
	cc      resolver.ClientConn           // client connection
	watchCh clientv3.WatchChan            // watch channel
	closeCh chan struct{}                 // close channel
	store   map[string]mapset.Set[string] // store key:srv/endpint v:endpoint
}

func NewResolver() *Resolver {
	return new(Resolver)
}

func (r *Resolver) Scheme() string {
	return Scheme
}

func (r *Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   Endpoints,
		DialTimeout: time.Second * time.Duration(TTL),
	})
	if err != nil {
		return nil, errors.Wrap(err, "client new")
	}

	r = &Resolver{
		client:  client,
		target:  target,
		cc:      cc,
		closeCh: make(chan struct{}, 1),
		store:   make(map[string]mapset.Set[string]),
	}

	return r, r.start()
}

func (r *Resolver) ResolveNow(o resolver.ResolveNowOptions) {}

func (r *Resolver) Close() {
	close(r.closeCh)
}

func (r *Resolver) start() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(TTL))
	defer cancel()

	res, err := r.client.Get(ctx, r.target.URL.Path, clientv3.WithPrefix())
	if err != nil {
		return errors.Wrap(err, "etcd get")
	}

	endpoints := mapset.NewSet[string]()

	addr := make([]resolver.Address, len(res.Kvs))
	for i, item := range res.Kvs {
		addr[i] = resolver.Address{
			Addr: string(item.Value),
		}

		endpoints.Add(string(item.Value))
	}
	r.cc.UpdateState(resolver.State{
		Addresses: addr,
	})

	r.store[r.target.URL.Path] = endpoints

	go r.watch()
	return nil
}

func (r *Resolver) watch() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	w := clientv3.NewWatcher(r.client)
	defer w.Close()

	endpoints := r.store[r.target.URL.Path]
	r.watchCh = w.Watch(context.Background(), r.target.URL.Path, clientv3.WithPrefix())
	for {
		select {
		case res := <-r.watchCh:
			for _, ev := range res.Events {
				switch ev.Type {
				case mvccpb.PUT:
					v := string(ev.Kv.Value)
					if !endpoints.Contains(v) {
						endpoints.Add(v)
					}
				case mvccpb.DELETE:
					k := string(ev.Kv.Key)
					v := filepath.Base(k)
					if endpoints.Contains(v) {
						endpoints.Remove(v)
					}
				}
			}

			addr := make([]resolver.Address, endpoints.Cardinality())
			for i, item := range endpoints.ToSlice() {
				addr[i] = resolver.Address{
					Addr: item,
				}
			}
			r.cc.UpdateState(resolver.State{
				Addresses: addr,
			})
		case <-ticker.C:
			r.ResolveNow(resolver.ResolveNowOptions{})
		case <-r.closeCh:
			r.client.Close()
			return
		}
	}
}
