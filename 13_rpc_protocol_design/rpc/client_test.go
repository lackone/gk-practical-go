package rpc

import (
	"context"
	"fmt"
	"gk-practical-go/13_rpc_protocol_design/proto/gen"
	"gk-practical-go/13_rpc_protocol_design/rpc/compress"
	"gk-practical-go/13_rpc_protocol_design/rpc/serialize"
	"testing"
	"time"
)

func TestClientProto(t *testing.T) {
	us := &UserService{}
	client, _ := NewClient(":8888", WithSerializer(&serialize.Proto{}))
	client.InitClientProxy(us)

	resp, err := us.GetByIdProto(context.Background(), &gen.GetByIdReq{
		Id: 111,
	})
	if err != nil {
		fmt.Printf("%#v\n", err)
	}
	fmt.Printf("%#v\n", resp)
	fmt.Printf("%#v\n", resp.User)
}

func TestClient(t *testing.T) {
	us := &UserService{}

	client, _ := NewClient(":8888")
	client.InitClientProxy(us)

	resp, err := us.GetById(context.Background(), &GetByIdReq{
		Id: 111,
	})
	if err != nil {
		fmt.Printf("%#v\n", err)
	}
	fmt.Printf("%#v\n", resp)
}

func TestClientOneway(t *testing.T) {
	us := &UserService{}

	client, _ := NewClient(":8888")
	client.InitClientProxy(us)

	resp, err := us.GetById(CtxWithOneway(context.Background()), &GetByIdReq{
		Id: 111,
	})
	if err != nil {
		fmt.Printf("%#v\n", err)
	}
	fmt.Printf("%#v\n", resp)
}

func TestClientTimeout(t *testing.T) {
	us := &UserService{}

	client, _ := NewClient(":8888")
	client.InitClientProxy(us)

	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	resp, err := us.GetById(ctx, &GetByIdReq{
		Id: 111,
	})
	if err != nil {
		fmt.Printf("%#v\n", err)
	}
	fmt.Printf("%#v\n", resp)
}

func TestClientGzip(t *testing.T) {
	us := &UserService{}
	client, _ := NewClient(":8888", WithCompressor(&compress.Gzip{}))
	client.InitClientProxy(us)

	resp, err := us.GetById(context.Background(), &GetByIdReq{
		Id:   111,
		Data: "aaa bbb ccc a b c 123",
	})
	if err != nil {
		fmt.Printf("%#v\n", err)
	}
	fmt.Printf("%#v\n", resp)
}
