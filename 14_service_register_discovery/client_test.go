package _4_service_register_discovery

import (
	"context"
	"fmt"
	"gk-practical-go/14_service_register_discovery/proto/gen"
	"gk-practical-go/14_service_register_discovery/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
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

	client, err := NewClient(WithInsecure(), WithRegistryRegistry(reg), WithTimeout(time.Second*3))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)

	cc, err := client.Dial(ctx, "user-service")
	if err != nil {
		t.Log(err)
	}

	defer cancel()

	uc := gen.NewUserServiceClient(cc)

	resp, err := uc.GetById(ctx, &gen.GetByIdReq{
		Id: 123,
	})
	if err != nil {
		t.Log(err)
	}

	fmt.Println(resp)
}
