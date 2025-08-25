package rpc

import (
	"context"
	"fmt"
	"gk-practical-go/13_rpc_protocol_design/proto/gen"
	"time"
)

type UserService struct {
	GetById func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)

	GetByIdProto func(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error)
}

func (u *UserService) Name() string {
	return "UserService"
}

type GetByIdReq struct {
	Id   int
	Data string
}

type GetByIdResp struct {
	Msg string
}

type UserServiceServer struct {
	Err error
	Msg string
}

func (u *UserServiceServer) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	fmt.Println(req)
	fmt.Println(req.Data)
	return &GetByIdResp{Msg: u.Msg}, u.Err
}

func (u *UserServiceServer) GetByIdProto(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	fmt.Println(req)
	return &gen.GetByIdResp{
		User: &gen.User{
			Id:   req.Id,
			Name: u.Msg,
		},
	}, u.Err
}

func (u *UserServiceServer) Name() string {
	return "UserService"
}

type UserServiceServerTimeout struct {
	Timeout time.Duration
	Err     error
	Msg     string
}

func (u *UserServiceServerTimeout) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	fmt.Println(req)
	time.Sleep(u.Timeout)
	return &GetByIdResp{Msg: u.Msg}, u.Err
}

func (u *UserServiceServerTimeout) GetByIdProto(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	fmt.Println(req)
	time.Sleep(u.Timeout)
	return &gen.GetByIdResp{
		User: &gen.User{
			Id:   req.Id,
			Name: u.Msg,
		},
	}, u.Err
}

func (u *UserServiceServerTimeout) Name() string {
	return "UserService"
}
