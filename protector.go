package protector

import (
	"log"
	"time"

	"github.com/coreos/etcd/clientv3"
)

var (
	// Endpoints Endpoints
	Endpoints = []string{
		"http://127.0.0.1:2379",
		"http://127.0.0.1:22379",
		"http://127.0.0.1:32379",
	}
	// DialTimeout DialTimeout
	DialTimeout = 5 * time.Second
	// Interval Interval
	Interval = 10 * time.Second
	// TTL ttl
	TTL int64 = 15
	// Prefix prefix
	Prefix = "rpcx.io"
)

// Protector protector
type Protector struct{}

// NewProtector new protector
func NewProtector() *Protector {
	return &Protector{}
}

// Client etcd client
func (p *Protector) Client() *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   Endpoints,
		DialTimeout: DialTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	return cli
}
