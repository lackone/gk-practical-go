package random

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math/rand"
	"strconv"
)

type WeightBalancer struct {
	conns       []*conn
	len         int
	totalWeight uint32
}

func (b *WeightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if b.len == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	tgt := rand.Intn(int(b.totalWeight) + 1)
	var idx int
	for i, c := range b.conns {
		tgt = tgt - int(c.weight)
		if tgt <= 0 {
			idx = i
		}
	}

	return balancer.PickResult{
		SubConn: b.conns[idx].c,
	}, nil
}

type WeightBuilder struct {
}

func (b *WeightBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*conn, len(info.ReadySCs))
	var totalWeight uint32
	for sb, sbInfo := range info.ReadySCs {
		weightStr, ok := sbInfo.Address.Attributes.Value("weight").(string)
		if !ok || weightStr == "" {
			//panic()
			//continue
		}
		weight, err := strconv.ParseUint(weightStr, 10, 32)
		if err != nil {
			panic(err)
		}

		totalWeight += uint32(weight)

		conns = append(conns, &conn{
			c:      sb,
			weight: uint32(weight),
		})
	}
	return &WeightBalancer{
		conns:       conns,
		len:         len(conns),
		totalWeight: totalWeight,
	}
}

type conn struct {
	c      balancer.SubConn
	weight uint32
}
