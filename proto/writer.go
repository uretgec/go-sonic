package proto

import (
	"io"
	"strings"
)

type writer interface {
	io.Writer
}

type Writer struct {
	writer
}

func NewWriter(wr writer) *Writer {
	return &Writer{
		writer: wr,
	}
}

func (w *Writer) WriteArgs(args []string) error {
	_, err := w.writer.Write([]byte(strings.Join(args, " ") + "\r\n"))
	return err
}
