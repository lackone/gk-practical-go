package _2_network_min_rpc

import (
	"encoding/binary"
	"io"
	"net"
)

func Serve() error {
	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			continue
		}

		go func() {
			if err := handleConn(conn); err != nil {
				conn.Close()
			}
		}()
	}
}

func handleConn(conn net.Conn) error {
	for {
		buf := make([]byte, 8)
		_, err := conn.Read(buf)
		if err == io.EOF || err == net.ErrClosed || err == io.ErrUnexpectedEOF {
			return err
		}
		if err != nil {
			continue
		}
		//if n != 8 {
		//	return errors.New("没有读够")
		//}
		res := handleMsg(buf)
		_, err = conn.Write(res)
		if err == io.EOF || err == net.ErrClosed || err == io.ErrUnexpectedEOF {
			return err
		}
		if err != nil {
			continue
		}
		//if n != len(res) {
		//	return errors.New("没有写完")
		//}
	}
}

func handleMsg(bs []byte) []byte {
	res := make([]byte, 2*len(bs))
	copy(res[:len(bs)], bs)
	copy(res[len(bs):], bs)
	return res
}

type Server struct {
	network string
	addr    string
	dataLen int
}

func NewServer(network, addr string) *Server {
	return &Server{
		network: network,
		addr:    addr,
		dataLen: 8,
	}
}

func (s *Server) Start() error {
	listen, err := net.Listen(s.network, s.addr)
	if err != nil {
		return err
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			continue
		}

		go func() {
			if err := s.handleConn(conn); err != nil {
				conn.Close()
			}
		}()
	}
}

func (s *Server) handleConn(conn net.Conn) error {
	for {
		dataLen := make([]byte, s.dataLen)
		_, err := conn.Read(dataLen)
		if err != nil {
			return err
		}

		//消息有多长
		length := binary.BigEndian.Uint64(dataLen)

		body := make([]byte, length)

		_, err = conn.Read(body)
		if err != nil {
			return err
		}

		resBody := handleMsg(body)
		resBodyLen := len(resBody)

		//在这构建响应数据
		res := make([]byte, resBodyLen+s.dataLen)

		//把长度写进去8个字节
		binary.BigEndian.PutUint64(res[:s.dataLen], uint64(resBodyLen))
		//写入数据
		copy(res[s.dataLen:], resBody)

		_, err = conn.Write(res)
		if err != nil {
			return err
		}
	}
}
