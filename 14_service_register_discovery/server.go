package _4_service_register_discovery

import (
	"context"
	"gk-practical-go/14_service_register_discovery/registry"
	"google.golang.org/grpc"
	"net"
	"time"
)

type Server struct {
	name            string
	registry        registry.Registry
	registerTimeout time.Duration
	*grpc.Server
	ls net.Listener
}

type ServerOption func(*Server)

func NewServer(name string, opts ...ServerOption) (*Server, error) {
	s := &Server{
		name:            name,
		Server:          grpc.NewServer(),
		registerTimeout: time.Second * 3,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

func WithRegistry(reg registry.Registry) ServerOption {
	return func(s *Server) {
		s.registry = reg
	}
}

func (s *Server) Start(addr string) error {
	ls, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.ls = ls

	if s.registry != nil {
		ctx, cancel := context.WithTimeout(context.Background(), s.registerTimeout)
		defer cancel()
		err := s.registry.Register(ctx, registry.ServiceInstance{
			Name: s.name,
			Addr: ls.Addr().String(),
		})
		if err != nil {
			return err
		}
		//注册成功
		//defer func() {
		//	s.registry.Close()
		//}()
	}

	err = s.Serve(ls)

	return err
}

func (s *Server) Close() error {
	if s.registry != nil {
		err := s.registry.Close()
		if err != nil {
			return err
		}
	}

	//err := s.ls.Close()
	s.Server.GracefulStop()
	return nil
}
