package message

import (
	"bytes"
	"encoding/binary"
)

type Request struct {
	//头消息长度
	HeadLen uint32
	//消息体长度
	BodyLen uint32
	//消息ID
	MsgID uint32
	//版本，一个字节
	Version uint8
	//压缩算法，一个字节
	Compress uint8
	//序列化，一个字节
	Serializer uint8
	//服务名
	ServiceName string
	//方法名
	MethodName string

	//扩展字段，用于传递自定义元数据
	Meta map[string]string

	//消息体
	Data []byte
}

func (r *Request) CalcHeadLen() {
	headLen := 15 + len(r.ServiceName) + 1 + len(r.MethodName) + 1
	for k, v := range r.Meta {
		headLen += len(k) + 1
		headLen += len(v) + 1
	}
	r.HeadLen = uint32(headLen)
}

func (r *Request) CalcBodyLen() {
	r.BodyLen = uint32(len(r.Data))
}

func EncodeReq(req *Request) []byte {
	body := make([]byte, req.HeadLen+req.BodyLen)

	//写入头部长度
	binary.BigEndian.PutUint32(body[0:4], req.HeadLen)
	//写入体长度
	binary.BigEndian.PutUint32(body[4:8], req.BodyLen)
	//写入消息ID
	binary.BigEndian.PutUint32(body[8:12], req.MsgID)

	body[12] = req.Version
	body[13] = req.Compress
	body[14] = req.Serializer

	cur := body[15:]
	copy(cur, req.ServiceName)
	cur = cur[len(req.ServiceName):]
	// 分隔符
	cur[0] = '\n'
	cur = cur[1:]
	copy(cur, req.MethodName)
	cur = cur[len(req.MethodName):]
	cur[0] = '\n'
	cur = cur[1:]

	for k, v := range req.Meta {
		copy(cur, k)
		cur = cur[len(k):]
		cur[0] = '\r'
		cur = cur[1:]
		copy(cur, v)
		cur = cur[len(v):]
		cur[0] = '\n'
		cur = cur[1:]
	}

	copy(cur, req.Data)

	return body
}

func DecodeReq(data []byte) *Request {
	req := &Request{}

	//头部长度
	req.HeadLen = binary.BigEndian.Uint32(data[0:4])
	//体长度
	req.BodyLen = binary.BigEndian.Uint32(data[4:8])
	//消息ID
	req.MsgID = binary.BigEndian.Uint32(data[8:12])

	req.Version = uint8(data[12])
	req.Compress = uint8(data[13])
	req.Serializer = uint8(data[14])

	header := data[15:req.HeadLen]

	index := bytes.IndexByte(header, '\n')

	req.ServiceName = string(header[:index])

	header = header[index+1:]

	index = bytes.IndexByte(header, '\n')

	req.MethodName = string(header[:index])

	header = header[index+1:]

	//解析 meta
	index = bytes.IndexByte(header, '\n')
	if index != -1 {
		meta := make(map[string]string)

		for index != -1 {
			pair := header[:index]
			//再找到 \r
			pairIndex := bytes.IndexByte(pair, '\r')
			key := string(pair[:pairIndex])
			val := string(pair[pairIndex+1:])

			meta[key] = val

			header = header[index+1:]

			index = bytes.IndexByte(header, '\n')
		}

		req.Meta = meta
	}

	if req.BodyLen != 0 {
		req.Data = data[req.HeadLen:]
	}

	return req
}
