package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"sync/atomic"
)

type Balancer struct {
	index int32
	conns []balancer.SubConn
	len   int32
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	idx := atomic.AddInt32(&b.index, 1)
	c := b.conns[idx%b.len]

	return balancer.PickResult{
		SubConn: c,
	}, nil
}

type Builder struct {
}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]balancer.SubConn, len(info.ReadySCs))
	for c := range info.ReadySCs {
		conns = append(conns, c)
	}
	return &Balancer{
		conns: conns,
		index: -1,
		len:   int32(len(conns)),
	}
}
