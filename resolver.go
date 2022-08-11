package protector

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

type Resolver struct {
	client *clientv3.Client
	target resolver.Target
	cc     resolver.ClientConn
	store  map[string]struct{}
	stopCh chan struct{}
	rn     chan struct{} // rn channel is used by ResolveNow() to force an immediate resolution of the target.
	t      *time.Ticker
}

func (r *Resolver) watch(ctx context.Context) {
	target := fmt.Sprintf("%s/", r.target.URL.Path)

	w := clientv3.NewWatcher(r.client)
	rch := w.Watch(ctx, target, clientv3.WithPrefix())
	for {
		select {
		case <-r.rn:
			r.resolveNow()
		case <-r.t.C:
			r.ResolveNow(resolver.ResolveNowOptions{})
		case <-r.stopCh:
			w.Close()
			return
		case wresp := <-rch:
			for _, ev := range wresp.Events {
				switch ev.Type {
				case mvccpb.PUT:
					r.store[string(ev.Kv.Value)] = struct{}{}
				case mvccpb.DELETE:
					delete(r.store, strings.Replace(string(ev.Kv.Key), target, "", 1))
				}
			}
			r.updateTargetState()
		}
	}
}

func (r *Resolver) resolveNow() {
	target := fmt.Sprintf("%s/", r.target.URL.Path)

	resp, err := r.client.Get(context.Background(), target, clientv3.WithPrefix())
	if err != nil {
		r.cc.ReportError(errors.Wrap(err, "get init endpoints"))
		return
	}

	for _, kv := range resp.Kvs {
		r.store[string(kv.Value)] = struct{}{}
	}

	r.updateTargetState()
}

func (r *Resolver) updateTargetState() {
	addrs := make([]resolver.Address, len(r.store))

	i := 0
	for k := range r.store {
		addrs[i] = resolver.Address{Addr: k}
		i++
	}

	r.cc.UpdateState(resolver.State{Addresses: addrs})
}

// ResolveNow will be called by gRPC to try to resolve the target name
// again. It's just a hint, resolver can ignore this if it's not necessary.
//
// It could be called multiple times concurrently.
func (r *Resolver) ResolveNow(o resolver.ResolveNowOptions) {
	select {
	case r.rn <- struct{}{}:
	default:

	}
}

// Close closes the resolver.
func (r *Resolver) Close() {
	r.t.Stop()
	close(r.stopCh)
}
