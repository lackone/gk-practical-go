package _4_service_register_discovery

import (
	"context"
	"fmt"
	"gk-practical-go/14_service_register_discovery/proto/gen"
	"gk-practical-go/14_service_register_discovery/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
	"testing"
)

type ServerService struct {
	gen.UnimplementedUserServiceServer
}

func (s *ServerService) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	fmt.Println(req)
	return &gen.GetByIdResp{
		User: &gen.User{
			Id:   req.Id,
			Name: "test",
		},
	}, nil
}

func TestServer(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"http://192.168.1.222:2379"},
	})
	if err != nil {
		t.Log(err)
	}
	reg, err := etcd.NewRegistry(etcdClient)
	if err != nil {
		t.Log(err)
	}
	ss := &ServerService{}
	gs, err := NewServer("user-service", WithRegistry(reg))
	if err != nil {
		t.Log(err)
	}
	gen.RegisterUserServiceServer(gs, ss)
	err = gs.Start(":9090")
	if err != nil {
		t.Log(err)
	}
}
