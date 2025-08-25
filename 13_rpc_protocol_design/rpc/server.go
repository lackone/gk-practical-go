package rpc

import (
	"context"
	"errors"
	"fmt"
	"gk-practical-go/13_rpc_protocol_design/rpc/compress"
	"gk-practical-go/13_rpc_protocol_design/rpc/message"
	"gk-practical-go/13_rpc_protocol_design/rpc/serialize"
	"net"
	"reflect"
	"strconv"
	"time"
)

type Server struct {
	network string
	addr    string
	dataLen int

	services    map[string]ReflectionStub
	serializers map[uint8]serialize.Serializer
	compressors map[uint8]compress.Compressor
}

func NewServer(network, addr string) *Server {
	s := &Server{
		network:     network,
		addr:        addr,
		dataLen:     8,
		services:    make(map[string]ReflectionStub),
		serializers: make(map[uint8]serialize.Serializer),
		compressors: make(map[uint8]compress.Compressor),
	}
	s.RegisterSerializer(&serialize.Json{})
	s.RegisterCompressor(&compress.Nothing{})
	return s
}

func (s *Server) RegisterSerializer(ss serialize.Serializer) {
	s.serializers[ss.Code()] = ss
}

func (s *Server) RegisterCompressor(cc compress.Compressor) {
	s.compressors[cc.Code()] = cc
}

func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = ReflectionStub{
		s:           service,
		value:       reflect.ValueOf(service),
		serializers: s.serializers,
		compressors: s.compressors,
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

		req := message.DecodeReq(body)

		ctx := context.Background()
		oneway, ok := req.Meta["oneway"]

		if ok && oneway == "true" {
			ctx = CtxWithOneway(ctx)
		}

		cancel := func() {}

		if deadlineStr, ok := req.Meta["deadline"]; ok {
			if deadline, er := strconv.ParseInt(deadlineStr, 10, 64); er == nil {
				ctx, cancel = context.WithDeadline(ctx, time.UnixMilli(deadline))
			}
		}

		resp, err := s.Invoke(ctx, req)

		cancel()

		if err != nil {
			//暂时不知道怎么回传客户端
			resp.Error = []byte(err.Error())
		}

		resp.CalcHeadLen()
		resp.CalcBodyLen()

		_, err = conn.Write(message.EncodeResp(resp))
		if err != nil {
			return err
		}
	}
}

func (s *Server) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	resp := &message.Response{
		MsgID:      req.MsgID,
		Version:    req.Version,
		Compress:   req.Compress,
		Serializer: req.Serializer,
	}

	//还原了调用信息，你已经知道 service name, method name 和 参数了
	//要发起业务调用了
	service, ok := s.services[req.ServiceName]
	if !ok {
		return resp, errors.New("service not found")
	}

	if isOneway(ctx) {
		go func() {
			_, _ = service.invoke(ctx, req)
		}()
		return nil, errors.New("oneway")
	}

	res, err := service.invoke(ctx, req)
	resp.Data = res

	if err != nil {
		return resp, err
	}

	return resp, err
}

type ReflectionStub struct {
	s           Service
	value       reflect.Value
	serializers map[uint8]serialize.Serializer
	compressors map[uint8]compress.Compressor
}

func (s *ReflectionStub) invoke(ctx context.Context, req *message.Request) ([]byte, error) {

	method := s.value.MethodByName(req.MethodName)

	in := make([]reflect.Value, 2)

	//参数1
	in[0] = reflect.ValueOf(ctx)

	//解压缩
	compress, ok := s.compressors[req.Compress]
	if !ok {
		return nil, errors.New("compress not found")
	}

	fmt.Println(string(req.Data))

	data, err := compress.UnCompress(req.Data)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(data))

	//反顺列化
	serializer, ok := s.serializers[req.Serializer]
	if !ok {
		return nil, errors.New("serializer not found")
	}

	inArg := reflect.New(method.Type().In(1).Elem())
	err = serializer.Decode(data, inArg.Interface())

	if err != nil {
		return nil, err
	}

	//参数2
	in[1] = inArg

	//调用方法
	ret := method.Call(in)

	//ret[0] 是返回值
	//ret[1] 是error
	if ret[1].Interface() != nil {
		err = ret[1].Interface().(error)
	}

	var res []byte

	if ret[0].IsNil() {
		return nil, err
	} else {
		var er error
		res, er = serializer.Encode(ret[0].Interface())
		if er != nil {
			return nil, er
		}
		res, er = compress.Compress(res)
		if er != nil {
			return nil, er
		}
	}

	return res, err
}
