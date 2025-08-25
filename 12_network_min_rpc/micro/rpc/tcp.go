package rpc

import (
	"encoding/binary"
	"net"
)

func ReadMsg(conn net.Conn, len int) ([]byte, error) {
	dataLen := make([]byte, len)
	_, err := conn.Read(dataLen)
	if err != nil {
		return nil, err
	}

	//消息有多长
	length := binary.BigEndian.Uint64(dataLen)

	body := make([]byte, length)

	_, err = conn.Read(body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func EncodeMsg(data []byte, dataLen int) []byte {
	resBodyLen := len(data)

	//在这构建响应数据
	res := make([]byte, resBodyLen+dataLen)

	//把长度写进去8个字节
	binary.BigEndian.PutUint64(res[:dataLen], uint64(resBodyLen))
	//写入数据
	copy(res[dataLen:], data)

	return res
}
