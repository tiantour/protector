package protector

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func Register(service, host, port string) error {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   Endpoints,
		DialTimeout: time.Second * time.Duration(TTL),
	})
	if err != nil {
		return errors.Wrap(err, "new client")
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res, err := client.Grant(ctx, int64(TTL))
	if err != nil {
		return errors.Wrap(err, "etcd grant")
	}

	endpoint := fmt.Sprintf("%s%s", host, port)
	target := fmt.Sprintf("/%s/%s", service, endpoint)
	_, err = client.Put(ctx, target, endpoint, clientv3.WithLease(res.ID))
	if err != nil {
		return errors.Wrap(err, "etcd put")
	}

	ch, err := client.KeepAlive(ctx, res.ID)
	if err != nil {
		return errors.Wrap(err, "etcd keep alive")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case ka := <-ch:
			if ka != nil {
				continue
			}
			return Register(service, host, port)
		}
	}
}
