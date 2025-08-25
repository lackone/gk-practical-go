package _4_service_register_discovery

import (
	"context"
	"fmt"
	"gk-practical-go/14_service_register_discovery/registry"
	"google.golang.org/grpc/resolver"
	"time"
)

type GrpcBuilder struct {
	r       registry.Registry
	timeout time.Duration
}

func NewGrpcBuilder(r registry.Registry, t time.Duration) *GrpcBuilder {
	return &GrpcBuilder{
		r:       r,
		timeout: t,
	}
}

func (b *GrpcBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {

	r := &GrpcResolver{
		cc:      cc,
		r:       b.r,
		target:  target,
		timeout: b.timeout,
	}

	r.resolve()

	go r.watch()

	return r, nil
}

func (b *GrpcBuilder) Scheme() string {
	return "registry"
}

type GrpcResolver struct {
	target  resolver.Target
	cc      resolver.ClientConn
	r       registry.Registry
	timeout time.Duration
	close   chan struct{}
}

func (r *GrpcResolver) ResolveNow(options resolver.ResolveNowOptions) {
	r.resolve()
}

func (r *GrpcResolver) watch() {
	event, err := r.r.Subscribe(r.target.Endpoint())
	if err != nil {
		r.cc.ReportError(err)
		return
	}
	for {
		select {
		case <-event:
			r.resolve()
		case <-r.close:
			return
		}
	}
}

func (r *GrpcResolver) resolve() {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	instances, err := r.r.ListServices(ctx, r.target.Endpoint())
	if err != nil {
		r.cc.ReportError(err)
		return
	}
	address := make([]resolver.Address, 0, len(instances))
	for _, instance := range instances {
		address = append(address, resolver.Address{Addr: instance.Addr, ServerName: instance.Name})
	}
	fmt.Println(address)
	err = r.cc.UpdateState(resolver.State{Addresses: address})
	if err != nil {
		r.cc.ReportError(err)
		return
	}
}

func (r *GrpcResolver) Close() {
	close(r.close)
}
