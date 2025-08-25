package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"reflect"
)

type Server struct {
	network string
	addr    string
	dataLen int

	services map[string]ReflectionStub
}

func NewServer(network, addr string) *Server {
	return &Server{
		network:  network,
		addr:     addr,
		dataLen:  8,
		services: make(map[string]ReflectionStub),
	}
}

func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = ReflectionStub{
		s:     service,
		value: reflect.ValueOf(service),
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
		body, err := ReadMsg(conn, s.dataLen)
		if err != nil {
			return err
		}

		req := &Request{}

		err = json.Unmarshal(body, req)
		if err != nil {
			return err
		}

		resp, err := s.Invoke(context.Background(), req)

		if err != nil {
			//暂时不知道怎么回传客户端
		}

		res := EncodeMsg(resp.Data, s.dataLen)

		_, err = conn.Write(res)
		if err != nil {
			return err
		}
	}
}

func (s *Server) Invoke(ctx context.Context, req *Request) (*Response, error) {
	//还原了调用信息，你已经知道 service name, method name 和 参数了
	//要发起业务调用了
	service, ok := s.services[req.ServiceName]
	if !ok {
		return nil, errors.New("service not found")
	}

	res, err := service.invoke(context.Background(), req.MethodName, req.Arg)
	if err != nil {
		return nil, err
	}

	return &Response{
		Data: res,
	}, err
}

type ReflectionStub struct {
	s     Service
	value reflect.Value
}

func (s *ReflectionStub) invoke(ctx context.Context, methodName string, data []byte) ([]byte, error) {

	method := s.value.MethodByName(methodName)

	in := make([]reflect.Value, 2)

	//参数1
	in[0] = reflect.ValueOf(context.Background())

	//参数2
	inArg := reflect.New(method.Type().In(1).Elem())
	err := json.Unmarshal(data, inArg.Interface())

	if err != nil {
		return nil, err
	}

	in[1] = inArg

	//调用方法
	ret := method.Call(in)
	//ret[0] 是返回值
	//ret[1] 是error
	if ret[1].Interface() != nil {
		return nil, ret[1].Interface().(error)
	}

	return json.Marshal(ret[0].Interface())
}
