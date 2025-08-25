package rpc

import (
	"context"
	"errors"
	"github.com/silenceper/pool"
	"gk-practical-go/13_rpc_protocol_design/rpc/compress"
	"gk-practical-go/13_rpc_protocol_design/rpc/message"
	"gk-practical-go/13_rpc_protocol_design/rpc/serialize"
	"net"
	"reflect"
	"strconv"
	"time"
)

// InitClientProxy 要为 GetById 之类的函数类型的字段赋值
func (c *Client) InitClientProxy(service Service) error {
	return setFuncField(service, c, c.serializer, c.compressor)
}

func setFuncField(service Service, p Proxy, s serialize.Serializer, c compress.Compressor) error {
	if service == nil {
		return errors.New("rpc 不支持")
	}

	val := reflect.ValueOf(service)
	typ := val.Type()

	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return errors.New("rpc 只支持指向构体的一级指针")
	}

	valEl := val.Elem()
	typEl := typ.Elem()

	numField := typEl.NumField()
	for i := 0; i < numField; i++ {
		fieldType := typEl.Field(i)
		fieldVal := valEl.Field(i)

		if fieldVal.CanSet() {
			// 这个地方才是真正的将本地调用捕捉到的地方
			fn := func(args []reflect.Value) []reflect.Value {

				retVal := reflect.New(fieldType.Type.Out(0).Elem())

				//真的发起调用
				ctx := args[0].Interface().(context.Context)

				//序列化
				argData, err := s.Encode(args[1].Interface())
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				//压缩
				argData, err = c.Compress(argData)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				meta := make(map[string]string)

				if isOneway(ctx) {
					meta["oneway"] = "true"
				}

				if deadline, ok := ctx.Deadline(); ok {
					meta["deadline"] = strconv.FormatInt(deadline.UnixMilli(), 10)
				}

				req := &message.Request{
					ServiceName: service.Name(),
					MethodName:  fieldType.Name,
					//Args: MapConvert[reflect.Value, any](args, func(idx int, src reflect.Value) any {
					//	return src.Interface()
					//}),
					Data:       argData,
					Serializer: s.Code(),
					Compress:   c.Code(),
					Meta:       meta,
				}

				req.CalcHeadLen()
				req.CalcBodyLen()

				resp, err := p.Invoke(ctx, req)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				var serverErr error
				if len(resp.Error) > 0 {
					// 服务端传过来的 ERROR
					serverErr = errors.New(string(resp.Error))
				}

				if len(resp.Data) > 0 {
					data, err := c.UnCompress(resp.Data)
					if err != nil {
						return []reflect.Value{retVal, reflect.ValueOf(err)}
					}
					err = s.Decode(data, retVal.Interface())
					if err != nil {
						return []reflect.Value{retVal, reflect.ValueOf(err)}
					}
				}

				var serverErrVal reflect.Value
				if serverErr == nil {
					serverErrVal = reflect.Zero(reflect.TypeOf(new(error)).Elem())
				} else {
					serverErrVal = reflect.ValueOf(serverErr)
				}

				return []reflect.Value{retVal, serverErrVal}
			}

			//我要设置值给 GetById
			fnVal := reflect.MakeFunc(fieldType.Type, fn)

			fieldVal.Set(fnVal)
		}
	}

	return nil
}

type Client struct {
	network    string
	addr       string
	dataLen    int
	p          pool.Pool
	serializer serialize.Serializer
	compressor compress.Compressor
}

type ClientOption func(*Client)

func NewClient(addr string, opts ...ClientOption) (*Client, error) {
	p, err := pool.NewChannelPool(&pool.Config{
		InitialCap: 1,
		MaxCap:     30,
		MaxIdle:    10,
		Factory: func() (interface{}, error) {
			return net.DialTimeout("tcp", addr, time.Second*3)
		},
		Close: func(v interface{}) error {
			return v.(net.Conn).Close()
		},
	})
	if err != nil {
		return nil, err
	}
	res := &Client{
		network:    "tcp",
		addr:       addr,
		dataLen:    8,
		p:          p,
		serializer: &serialize.Json{},
		compressor: &compress.Nothing{},
	}

	for _, opt := range opts {
		opt(res)
	}

	return res, nil
}

func WithSerializer(s serialize.Serializer) ClientOption {
	return func(c *Client) {
		c.serializer = s
	}
}

func WithCompressor(cc compress.Compressor) ClientOption {
	return func(c *Client) {
		c.compressor = cc
	}
}

func (c *Client) send(ctx context.Context, data []byte) ([]byte, error) {
	val, err := c.p.Get()
	if err != nil {
		return nil, err
	}

	conn := val.(net.Conn)

	defer conn.Close()

	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}

	// 是否单向请求
	if isOneway(ctx) {
		return nil, err
	}

	return ReadMsg(conn, c.dataLen)
}

func (c *Client) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	ch := make(chan struct{})
	var resp *message.Response
	var err error

	go func() {
		resp, err = c.doInvoke(ctx, req)
		ch <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-ch:
		return resp, err
	}
}

func (c *Client) doInvoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	data := message.EncodeReq(req)

	//把请求发过去
	//正儿八经地把请求发过去服务端
	res, err := c.send(ctx, data)
	if err != nil {
		return nil, err
	}

	return message.DecodeResp(res), nil
}

func MapConvert[Src any, Dst any](src []Src, m func(idx int, src Src) Dst) []Dst {
	dst := make([]Dst, len(src))
	for i, s := range src {
		dst[i] = m(i, s)
	}
	return dst
}
