package last_active

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"sync/atomic"
)

type Balancer struct {
	conns []*conn
	len   int
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if b.len == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	var res *conn
	for _, c := range b.conns {
		if res == nil || c.cnt < res.cnt {
			res = c
		}
	}

	atomic.AddUint32(&res.cnt, 1)

	return balancer.PickResult{
		SubConn: res.c,
		Done: func(info balancer.DoneInfo) {
			atomic.AddUint32(&res.cnt, -1)
		},
	}, nil
}

type Builder struct {
}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*conn, len(info.ReadySCs))
	for sb := range info.ReadySCs {
		conns = append(conns, &conn{
			c: sb,
		})
	}
	return &Balancer{
		conns: conns,
		len:   len(conns),
	}
}

type conn struct {
	cnt uint32 //正在处理的请求
	c   balancer.SubConn
}
