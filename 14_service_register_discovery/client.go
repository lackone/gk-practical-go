package _4_service_register_discovery

import (
	"context"
	"fmt"
	"gk-practical-go/14_service_register_discovery/registry"
	"google.golang.org/grpc"
	"time"
)

type Client struct {
	insecure bool
	r        registry.Registry
	timeout  time.Duration
}

type ClientOption func(*Client)

func NewClient(opts ...ClientOption) (*Client, error) {
	res := &Client{}

	for _, opt := range opts {
		opt(res)
	}

	return res, nil
}

func WithInsecure() ClientOption {
	return func(c *Client) {
		c.insecure = true
	}
}

func WithTimeout(t time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = t
	}
}

func WithRegistryRegistry(r registry.Registry) ClientOption {
	return func(c *Client) {
		c.r = r
	}
}

func (c *Client) Dial(ctx context.Context, service string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{}
	if c.r != nil {
		rb := NewGrpcBuilder(c.r, c.timeout)
		opts = append(opts, grpc.WithResolvers(rb))
	}
	if c.insecure {
		opts = append(opts, grpc.WithInsecure())
	}
	cc, err := grpc.DialContext(ctx, fmt.Sprintf("registry:///%s", service), opts...)
	return cc, err
}
