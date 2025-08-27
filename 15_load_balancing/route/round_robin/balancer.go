package round_robin

import (
	"gk-practical-go/15_load_balancing/route"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"sync/atomic"
)

type Balancer struct {
	index  int32
	conns  []*conn
	len    int32
	filter route.Filter
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {

	tmp := make([]*conn, 0, len(b.conns))
	for _, conn := range b.conns {
		if !b.filter(info, conn.addr) {
			continue
		}
		tmp = append(tmp, conn)
	}

	idx := atomic.AddInt32(&b.index, 1)
	c := tmp[int(idx)%len(tmp)]

	return balancer.PickResult{
		SubConn: c.c,
	}, nil
}

type Builder struct {
	filter route.Filter
}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*conn, len(info.ReadySCs))
	for c, info := range info.ReadySCs {
		conns = append(conns, &conn{
			c:    c,
			addr: info.Address,
		})
	}

	filter := func(info balancer.PickInfo, addr resolver.Address) bool {
		return true
	}

	if b.filter != nil {
		filter = b.filter
	}

	return &Balancer{
		conns:  conns,
		index:  -1,
		len:    int32(len(conns)),
		filter: filter,
	}
}

type conn struct {
	c    balancer.SubConn
	addr resolver.Address
}
