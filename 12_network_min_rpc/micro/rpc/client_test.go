package rpc

import (
	"context"
	"testing"
)

func TestClient(t *testing.T) {
	us := &UserService{}
	InitClientProxy(":8888", us)

	resp, err := us.GetById(context.Background(), &GetByIdReq{
		Id: 111,
	})
	if err != nil {
		t.Log(err)
	}
	t.Log(resp)
}
