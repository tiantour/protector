package protector

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"golang.org/x/net/context"
)

// Register register
type Register struct {
	Client *clientv3.Client
	Traget string
	Addr   string
	Stop   chan bool
}

// NewRegister new register
func NewRegister() *Register {
	cli := NewProtector().Client()
	return &Register{
		Client: cli,
		Stop:   make(chan bool, 1),
	}
}

// Add add kv
func (r *Register) Add(srv, host, port string) error {
	r.Addr = fmt.Sprintf("%s%s", host, port)
	r.Traget = fmt.Sprintf("/%s/%s/%s", Prefix, srv, r.Addr)
	go func() {
		ticker := time.NewTicker(Interval)
		for {
			res, err := r.Client.Grant(context.Background(), TTL)
			if err != nil {
				log.Println(err)
			}
			_, err = r.Client.Get(context.Background(), r.Traget)
			if err != nil && err != rpctypes.ErrKeyNotFound {
				log.Println(err)
			}
			_, err = r.Client.Put(context.Background(), r.Traget, r.Addr, clientv3.WithLease(res.ID))
			if err != nil {
				log.Println("")
			}

			select {
			case <-r.Stop:
				return
			case <-ticker.C:
			}
		}
	}()
	return nil
}

// Delete delete kv
func (r *Register) Delete() error {
	r.Stop <- true
	r.Stop = make(chan bool, 1)
	_, err := r.Client.Delete(context.Background(), r.Traget)
	return err
}

// Server etcd server
func (r *Register) Server(srv, host, port string) error {
	err := r.Add(srv, host, port)
	if err != nil {
		return err
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		<-ch
		r.Delete()
		os.Exit(1)
	}()
	return nil
}
