package serialize

import "encoding/json"

type Json struct {
}

func (j *Json) Code() uint8 {
	return 1
}

func (j *Json) Encode(val any) ([]byte, error) {
	return json.Marshal(val)
}

func (j *Json) Decode(data []byte, val any) error {
	return json.Unmarshal(data, val)
}
