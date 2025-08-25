package rpc

import (
	"context"
	"fmt"
)

type UserService struct {
	GetById func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)
}

func (u *UserService) Name() string {
	return "UserService"
}

type GetByIdReq struct {
	Id int
}

type GetByIdResp struct {
	Msg string
}

type UserServiceServer struct {
}

func (u *UserServiceServer) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	fmt.Println(req)
	return &GetByIdResp{Msg: "Hello World"}, nil
}

func (u *UserServiceServer) Name() string {
	return "UserService"
}
