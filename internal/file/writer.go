package file

import "os"

type Writer struct {
	file *os.File
}

func NewWriter(path string) (*Writer, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return &Writer{
		file,
	}, nil
}

func (w *Writer) WriteChunk(chunk []byte) error {
	_, err := w.file.Write(chunk)
	if err != nil {
		return err
	}
	return nil
}

func (w *Writer) Close() error {
	return w.file.Close()
}
