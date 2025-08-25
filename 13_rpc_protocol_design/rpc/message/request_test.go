package message

import (
	"fmt"
	"testing"
)

func TestRequest(t *testing.T) {

	meta := map[string]string{
		"key": "value",
	}
	data := "hello world"

	hlen := uint32(15 + len("test") + 1 + len("test") + 1)
	for k, v := range meta {
		hlen += uint32(len(k) + 1 + len(v) + 1)
	}

	blen := uint32(len(data))

	req := &Request{
		HeadLen:     hlen,
		BodyLen:     blen,
		MsgID:       1,
		Version:     1,
		Compress:    1,
		Serializer:  1,
		ServiceName: "test",
		MethodName:  "test",
		Meta:        meta,
		Data:        []byte(data),
	}

	encodeReq := EncodeReq(req)
	fmt.Println(encodeReq)
	decodeReq := DecodeReq(encodeReq)
	fmt.Printf("%#v\n", req)
	fmt.Printf("%#v\n", decodeReq)
}
