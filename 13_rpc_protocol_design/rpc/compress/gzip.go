package compress

import (
	"bytes"
	"compress/gzip"
	"io"
)

type Gzip struct {
}

func (g *Gzip) Code() uint8 {
	return 1
}

func (g *Gzip) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	// 创建 gzip writer，写入 buf
	gzWriter := gzip.NewWriter(&buf)

	// 写入数据
	_, err := gzWriter.Write(data)
	if err != nil {
		return nil, err
	}

	// 必须调用 Close()，否则 gzip 数据不完整
	if err := gzWriter.Close(); err != nil {
		return nil, err
	}

	// 返回压缩后的数据
	return buf.Bytes(), nil
}

func (g *Gzip) UnCompress(data []byte) ([]byte, error) {
	// 将压缩数据放入 bytes.Reader
	buf := bytes.NewReader(data)

	// 创建 gzip 读取器
	gzReader, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err // 可能不是有效的 gzip 数据
	}
	defer gzReader.Close() // 记得关闭

	// 读取并解压所有数据
	originalData, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, err
	}

	return originalData, nil
}
