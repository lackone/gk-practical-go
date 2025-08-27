package round_robin

import (
	"context"
	"fmt"
	"gk-practical-go/14_service_register_discovery/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"net"
	"testing"
	"time"
)

type Server struct {
	gen.UnimplementedUserServiceServer
}

func (s Server) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	fmt.Println(req)
	return &gen.GetByIdResp{
		User: &gen.User{
			Id:   req.Id,
			Name: "test",
		},
	}, nil
}

func TestBalancer(t *testing.T) {
	go func() {
		//服务端
		s := &Server{}
		gs := grpc.NewServer()
		gen.RegisterUserServiceServer(gs, s)
		ls, err := net.Listen("tcp", ":9090")
		if err != nil {
			t.Log(err)
		}
		gs.Serve(ls)
	}()

	time.Sleep(1 * time.Second)

	balancer.Register(base.NewBalancerBuilder("DEMO_ROUND_ROBIN", &Builder{}, base.Config{HealthCheck: true}))

	//客户端
	cc, err := grpc.Dial("127.0.0.1:9090", grpc.WithInsecure(), grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`))
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
