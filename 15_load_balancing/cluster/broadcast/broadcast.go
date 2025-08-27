package broadcast

import (
	"context"
	"gk-practical-go/14_service_register_discovery/registry"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type ClusterBuilder struct {
	registry registry.Registry
	service  string
}

func (c ClusterBuilder) BuildUnaryInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if !isBroadCast(ctx) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		instances, err := c.registry.ListServices(ctx, c.service)
		if err != nil {
			return err
		}
		var eg errgroup.Group
		for _, instance := range instances {
			addr := instance.Addr
			eg.Go(func() error {
				insCC, err := grpc.Dial(addr)
				if err != nil {
					return err
				}
				//对每个节点进行调用
				err = invoker(ctx, method, req, reply, insCC, opts...)
				if err != nil {
					return err
				}
				return nil
			})
		}
		return eg.Wait()
	}
}

type broadcastKey struct {
}

func UseBroadCast(ctx context.Context) context.Context {
	return context.WithValue(ctx, broadcastKey{}, true)
}

func isBroadCast(ctx context.Context) bool {
	val, ok := ctx.Value(broadcastKey{}).(bool)
	return ok && val
}
