package _2_network_min_rpc

import "testing"

func TestClient(t *testing.T) {
	Connect("tcp", ":8080")
}

func TestClient2(t *testing.T) {
	client := NewClient("tcp", ":8888")

	send, err := client.send("hello")
	if err != nil {
		t.Log(err.Error())
	}
	t.Log(send)
}
