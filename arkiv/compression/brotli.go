package compression

import (
	"bytes"
	"fmt"
	"io"

	"github.com/andybalholm/brotli"
)

func BrotliCompress(data []byte) ([]byte, error) {

	if len(data) == 0 {
		return nil, nil
	}

	buf := bytes.NewBuffer(nil)

	writer := brotli.NewWriterV2(buf, 9)

	_, err := writer.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to write data to brotli compressor: %w", err)
	}
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close brotli compressor: %w", err)
	}

	return buf.Bytes(), nil

}

func MustBrotliCompress(data []byte) []byte {
	compressed, err := BrotliCompress(data)
	if err != nil {
		panic(fmt.Errorf("failed to compress data: %w", err))
	}
	return compressed
}

func BrotliDecompress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}
	reader := brotli.NewReader(bytes.NewReader(data))
	return io.ReadAll(reader)
}
