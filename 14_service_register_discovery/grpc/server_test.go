package grpc

import (
	"context"
	"fmt"
	"gk-practical-go/14_service_register_discovery/proto/gen"
	"google.golang.org/grpc"
	"net"
	"testing"
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

func TestServer(t *testing.T) {
	s := &Server{}
	gs := grpc.NewServer()
	gen.RegisterUserServiceServer(gs, s)
	ls, err := net.Listen("tcp", ":9090")
	if err != nil {
		t.Log(err)
	}
	gs.Serve(ls)
}
