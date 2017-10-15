# protector

grpc register ,  resolver , watcher fro etcd3

### conf

Endpoints Endpoints

	protector.Endpoints = []string{
		"http://127.0.0.1:2379",
		"http://127.0.0.1:22379",
		"http://127.0.0.1:32379",
	}

DialTimeout DialTimeout

	protector.DialTimeout = 5 * time.Second

Interval Interval

	protector.Interval = 10 * time.Second

 TTL ttl

	protector.TTL int64 = 15
	
// Prefix prefix

	protector.Prefix = "rpcx.io"

### server

register && wather
```
_ = protector.NewRegister().Server("hello_service", "127.0.0.1", ":50000")
	
```

### client

resolver

```
conn, err := protector.NewResolver().Client("hello_service")
if err != nil {
    panic(err)
}
```

### Tanks

[gRPC服务发现&负载均衡](http://colobu.com/2017/03/25/grpc-naming-and-load-balance/)