package protector

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	TimeToLive = 10
)

func Register(ctx context.Context, client *clientv3.Client, service, authority string) error {
	resp, err := client.Grant(ctx, TimeToLive)
	if err != nil {
		return errors.Wrap(err, "etcd grant")
	}

	_, err = client.Put(ctx, fmt.Sprintf("/%s/%s", service, authority), authority, clientv3.WithLease(resp.ID))
	if err != nil {
		return errors.Wrap(err, "etcd put")
	}

	respCh, err := client.KeepAlive(ctx, resp.ID)
	if err != nil {
		return errors.Wrap(err, "etcd keep alive")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-respCh:

		}
	}
}
