package rpc

import "testing"

func TestServer(t *testing.T) {
	server := NewServer("tcp", ":8888")
	server.RegisterService(&UserServiceServer{})
	server.Start()
}
