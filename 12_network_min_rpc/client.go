package _2_network_min_rpc

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func Connect(network, addr string) error {
	conn, err := net.DialTimeout(network, addr, time.Second*3)
	if err != nil {
		return err
	}
	defer conn.Close()
	for i := 0; i < 10; i++ {
		_, err := conn.Write([]byte("hello"))
		if err != nil {
			return err
		}
		res := make([]byte, 128)
		n, err := conn.Read(res)
		if err != nil {
			return err
		}
		fmt.Println(string(res[:n]))
	}
	return nil
}

type Client struct {
	network string
	addr    string
	dataLen int
}

func NewClient(network, addr string) *Client {
	return &Client{
		network: network,
		addr:    addr,
		dataLen: 8,
	}
}

func (c *Client) send(data string) (string, error) {
	conn, err := net.DialTimeout(c.network, c.addr, time.Second*3)
	if err != nil {
		return "", err
	}

	defer conn.Close()

	resBodyLen := len([]byte(data))

	//在这构建响应数据
	res := make([]byte, resBodyLen+c.dataLen)

	//把长度写进去8个字节
	binary.BigEndian.PutUint64(res[:c.dataLen], uint64(resBodyLen))
	//写入数据
	copy(res[c.dataLen:], []byte(data))

	_, err = conn.Write(res)
	if err != nil {
		return "", err
	}

	dataLen := make([]byte, c.dataLen)
	_, err = conn.Read(dataLen)
	if err != nil {
		return "", err
	}

	//消息有多长
	length := binary.BigEndian.Uint64(dataLen)

	body := make([]byte, length)

	_, err = conn.Read(body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
