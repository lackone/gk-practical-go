package message

import (
	"fmt"
	"testing"
)

func TestResponse(t *testing.T) {

	data := "hello world"

	hlen := uint32(15 + len("error"))

	blen := uint32(len(data))

	resp := &Response{
		HeadLen:    hlen,
		BodyLen:    blen,
		MsgID:      1,
		Version:    1,
		Compress:   1,
		Serializer: 1,
		Error:      []byte("error"),
		Data:       []byte(data),
	}

	encodeResp := EncodeResp(resp)
	fmt.Println(encodeResp)
	decodeResp := DecodeResp(encodeResp)
	fmt.Printf("%#v\n", resp)
	fmt.Printf("%#v\n", decodeResp)
}
