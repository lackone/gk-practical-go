package rpc

import (
	"encoding/binary"
	"net"
)

func ReadMsg(conn net.Conn, len int) ([]byte, error) {
	//协议头和协议体
	dataLen := make([]byte, len)

	_, err := conn.Read(dataLen)
	if err != nil {
		return nil, err
	}

	headLen := binary.BigEndian.Uint32(dataLen[:4])
	bodyLen := binary.BigEndian.Uint32(dataLen[4:])
	totalLen := headLen + bodyLen

	body := make([]byte, totalLen)

	_, err = conn.Read(body[8:])
	if err != nil {
		return nil, err
	}

	copy(body[:8], dataLen)

	return body, nil
}
