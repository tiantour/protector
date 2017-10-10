package protector

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc/naming"
)

// Watcher watcher
type Watcher struct {
	Client *clientv3.Client
	Target string
	Init   bool
}

// Close closes the Watcher.
func (w *Watcher) Close() {
}

// Next watcher next
func (w *Watcher) Next() ([]*naming.Update, error) {
	// get
	if !w.Init {
		res, err := w.Client.Get(context.Background(), w.Target, clientv3.WithPrefix())
		if err != nil {
			return []*naming.Update{}, err
		}
		w.Init = true
		data := []*naming.Update{}
		for _, v := range res.Kvs {
			if v.Value != nil {
				data = append(data, &naming.Update{
					Op:   naming.Add,
					Addr: string(v.Value),
				})
			}
			return data, nil
		}
	}
	// watch
	ch := w.Client.Watch(context.Background(), w.Target, clientv3.WithPrefix())
	for item := range ch {
		for _, v := range item.Events {
			switch v.Type {
			case mvccpb.PUT:
				return []*naming.Update{{
					Op:   naming.Add,
					Addr: string(v.Kv.Value),
				}}, nil
			case mvccpb.DELETE:
				return []*naming.Update{{
					Op:   naming.Delete,
					Addr: string(v.Kv.Value),
				}}, nil
			}
		}
	}
	return nil, nil
}
