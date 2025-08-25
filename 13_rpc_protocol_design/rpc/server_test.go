package rpc

import (
	"errors"
	"gk-practical-go/13_rpc_protocol_design/rpc/compress"
	"gk-practical-go/13_rpc_protocol_design/rpc/serialize"
	"testing"
	"time"
)

func TestServerProto(t *testing.T) {
	server := NewServer("tcp", ":8888")
	service := &UserServiceServer{
		Msg: "OK",
		Err: errors.New("service error"),
	}
	server.RegisterService(service)
	server.RegisterSerializer(&serialize.Proto{})
	server.Start()
}

func TestServer(t *testing.T) {
	server := NewServer("tcp", ":8888")
	service := &UserServiceServer{
		Msg: "OK",
		Err: errors.New("service error"),
	}
	server.RegisterService(service)
	server.Start()
}

func TestServerTimeout(t *testing.T) {
	server := NewServer("tcp", ":8888")
	service := &UserServiceServerTimeout{
		Timeout: time.Second * 3,
		Msg:     "OK",
		Err:     errors.New("service error"),
	}
	server.RegisterService(service)
	server.Start()
}

func TestServerGzip(t *testing.T) {
	server := NewServer("tcp", ":8888")
	service := &UserServiceServer{
		Msg: "OK aaaa bbbb cccc 123",
		Err: errors.New("service error"),
	}
	server.RegisterService(service)
	server.RegisterCompressor(&compress.Gzip{})
	server.Start()
}
