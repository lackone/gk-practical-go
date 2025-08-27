package random

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math/rand"
)

type Balancer struct {
	conns []balancer.SubConn
	len   int
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if b.len == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	idx := rand.Intn(b.len)

	return balancer.PickResult{
		SubConn: b.conns[idx],
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
		len:   len(conns),
	}
}
