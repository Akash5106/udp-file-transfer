package file

import (
	"io"
	"os"
)

type Reader struct {
	file      *os.File
	chunkSize int
}

func NewReader(path string, chunkSize int) (*Reader, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &Reader{
		reader,
		chunkSize,
	}, nil
}

func (r *Reader) NextChunk() ([]byte, error) {
	buffer := make([]byte, r.chunkSize)
	n, err := r.file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if n == 0 {
		return nil, io.EOF
	}
	return buffer[:n], nil
}

func (r *Reader) CloseFile() error {
	return r.file.Close()
}
