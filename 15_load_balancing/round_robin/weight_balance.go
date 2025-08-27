package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
)

type WeightBalancer struct {
	conns []*conn
	mutex sync.Mutex
}

func (b *WeightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var totalWeight uint32
	var res *conn

	b.mutex.Lock()

	for _, c := range b.conns {
		totalWeight += c.effWeight
		c.curWeight += c.effWeight
		if res == nil || res.curWeight < c.curWeight {
			res = c
		}
	}
	res.curWeight -= totalWeight

	b.mutex.Unlock()

	return balancer.PickResult{
		SubConn: res.c,
		Done: func(info balancer.DoneInfo) {
			weight := atomic.LoadUint32(&res.effWeight)
			if info.Err != nil && res.weight == 0 {
				return
			}
			if info.Err == nil && res.weight == math.MaxUint32 {
				return
			}
			newWeight := weight
			if info.Err != nil {
				newWeight--
			} else {
				newWeight++
			}
			if atomic.CompareAndSwapUint32(&res.effWeight, weight, newWeight) {
				return
			}
		},
	}, nil
}

type WeightBuilder struct {
}

func (b *WeightBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*conn, len(info.ReadySCs))
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
		conns = append(conns, &conn{
			c:      sb,
			weight: uint32(weight),
		})
	}
	return &WeightBalancer{
		conns: conns,
		mutex: sync.Mutex{},
	}
}

// 每个实例都有三个值：weight（权重）、currentWeight（当前权重）、efficientWeight（有效权重）。
// efficientWeight 会根据调用结果动态调整。
// 每次挑选实例的时候，计算所有实例的 efficientWeight 来作为 totalWeight。
// 对于每一个实例，更新 currentWeight 为 currentWeight + efficientWeight。
// 挑选 currentWeight 最大的那个节点作为最终节点，并且更新它的 currentWeight 为 currentWeight-totalWeight。
type conn struct {
	c         balancer.SubConn
	weight    uint32
	curWeight uint32
	effWeight uint32
}
