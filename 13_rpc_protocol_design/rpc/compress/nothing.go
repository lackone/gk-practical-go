package compress

type Nothing struct {
}

func (n *Nothing) Code() uint8 {
	return 0
}

func (n *Nothing) Compress(data []byte) ([]byte, error) {
	return data, nil
}

func (n *Nothing) UnCompress(data []byte) ([]byte, error) {
	return data, nil
}
