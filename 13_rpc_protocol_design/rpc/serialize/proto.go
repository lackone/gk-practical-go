package serialize

import (
	"errors"
	"google.golang.org/protobuf/proto"
)

type Proto struct {
}

func (p *Proto) Code() uint8 {
	return 2
}

func (p *Proto) Encode(val any) ([]byte, error) {
	msg, ok := val.(proto.Message)
	if !ok {
		return nil, errors.New("proto message is not a proto.Message")
	}
	return proto.Marshal(msg)
}

func (p *Proto) Decode(data []byte, val any) error {
	msg, ok := val.(proto.Message)
	if !ok {
		return errors.New("proto message is not a proto.Message")
	}
	return proto.Unmarshal(data, msg)
}
