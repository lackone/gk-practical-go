package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/silenceper/pool"
	"net"
	"reflect"
	"time"
)

// InitClientProxy 要为 GetById 之类的函数类型的字段赋值
func InitClientProxy(addr string, service Service) error {
	client, err := NewClient(addr)
	if err != nil {
		return err
	}
	return setFuncField(service, client)
}

func setFuncField(service Service, p Proxy) error {
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

				argData, err := json.Marshal(args[1].Interface())
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				req := &Request{
					ServiceName: service.Name(),
					MethodName:  fieldType.Name,
					//Args: MapConvert[reflect.Value, any](args, func(idx int, src reflect.Value) any {
					//	return src.Interface()
					//}),
					Arg: argData,
				}

				//真的发起调用
				ctx := args[0].Interface().(context.Context)

				resp, err := p.Invoke(ctx, req)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				err = json.Unmarshal(resp.Data, retVal.Interface())
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				return []reflect.Value{retVal, reflect.Zero(reflect.TypeOf(new(error)).Elem())}
			}

			//我要设置值给 GetById
			fnVal := reflect.MakeFunc(fieldType.Type, fn)

			fieldVal.Set(fnVal)
		}
	}

	return nil
}

type Client struct {
	network string
	addr    string
	dataLen int
	p       pool.Pool
}

func NewClient(addr string) (*Client, error) {
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
	return &Client{
		network: "tcp",
		addr:    addr,
		dataLen: 8,
		p:       p,
	}, nil
}

func (c *Client) send(data []byte) ([]byte, error) {
	val, err := c.p.Get()
	if err != nil {
		return nil, err
	}

	conn := val.(net.Conn)

	defer conn.Close()

	res := EncodeMsg(data, c.dataLen)

	_, err = conn.Write(res)
	if err != nil {
		return nil, err
	}

	return ReadMsg(conn, c.dataLen)
}

func (c *Client) Invoke(ctx context.Context, req *Request) (*Response, error) {
	//把请求发过去
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	//正儿八经地把请求发过去服务端
	res, err := c.send(data)
	if err != nil {
		return nil, err
	}

	return &Response{
		Data: res,
	}, nil
}

func MapConvert[Src any, Dst any](src []Src, m func(idx int, src Src) Dst) []Dst {
	dst := make([]Dst, len(src))
	for i, s := range src {
		dst[i] = m(i, s)
	}
	return dst
}
