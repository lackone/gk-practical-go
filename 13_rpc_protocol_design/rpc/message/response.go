package message

import "encoding/binary"

type Response struct {
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

	//错误信息
	Error []byte

	Data []byte
}

func (r *Response) CalcHeadLen() {
	r.HeadLen = 15 + uint32(len(r.Error))
}

func (r *Response) CalcBodyLen() {
	r.BodyLen = uint32(len(r.Data))
}

func EncodeResp(resp *Response) []byte {
	body := make([]byte, resp.HeadLen+resp.BodyLen)

	//写入头部长度
	binary.BigEndian.PutUint32(body[0:4], resp.HeadLen)
	//写入体长度
	binary.BigEndian.PutUint32(body[4:8], resp.BodyLen)
	//写入消息ID
	binary.BigEndian.PutUint32(body[8:12], resp.MsgID)

	body[12] = resp.Version
	body[13] = resp.Compress
	body[14] = resp.Serializer

	cur := body[15:]
	copy(cur, resp.Error)
	cur = cur[len(resp.Error):]
	copy(cur, resp.Data)

	return body
}

func DecodeResp(data []byte) *Response {
	resp := &Response{}

	//头部长度
	resp.HeadLen = binary.BigEndian.Uint32(data[0:4])
	//体长度
	resp.BodyLen = binary.BigEndian.Uint32(data[4:8])
	//消息ID
	resp.MsgID = binary.BigEndian.Uint32(data[8:12])

	resp.Version = uint8(data[12])
	resp.Compress = uint8(data[13])
	resp.Serializer = uint8(data[14])

	if resp.HeadLen > 15 {
		resp.Error = data[15:resp.HeadLen]
	}

	if resp.BodyLen != 0 {
		resp.Data = data[resp.HeadLen:]
	}

	return resp
}
