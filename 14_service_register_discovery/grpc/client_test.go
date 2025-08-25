package grpc

import (
	"context"
	"fmt"
	"gk-practical-go/14_service_register_discovery/grpc_resolver"
	"gk-practical-go/14_service_register_discovery/proto/gen"
	grpc "google.golang.org/grpc"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	cc, err := grpc.Dial("registry:///127.0.0.1:9090", grpc.WithInsecure(), grpc.WithResolvers(&grpc_resolver.Builder{}))
	if err != nil {
		t.Log(err)
	}

	client := gen.NewUserServiceClient(cc)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	resp, err := client.GetById(ctx, &gen.GetByIdReq{
		Id: 123,
	})
	if err != nil {
		t.Log(err)
	}

	fmt.Println(resp)
}
