package route

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

// 返回值 true 就是留下，false 就是丢掉
type Filter func(info balancer.PickInfo, addr resolver.Address) bool

type GroupFilterBuilder struct {
	Group string
}

func (g *GroupFilterBuilder) Build() Filter {
	return func(info balancer.PickInfo, addr resolver.Address) bool {
		group := addr.Attributes.Value("group")
		return group == g.Group
	}
}
