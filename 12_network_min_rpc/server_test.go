package _2_network_min_rpc

import "testing"

func TestServer(t *testing.T) {
	Serve()
}

func TestServer2(t *testing.T) {
	server := NewServer("tcp", ":8888")

	err := server.Start()
	if err != nil {
		t.Log(err.Error())
	}
}
