package response

import (
	"fmt"
	"io"

	"github.com/remcous/bootdev_http/internal/headers"
)

type Writer struct {
	writer      io.Writer
	writerState writerState
}

type writerState int

const (
	writerStateStatusLine writerState = iota
	writerStateHeader
	writerStateBody
)

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer:      w,
		writerState: writerStateStatusLine,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != writerStateStatusLine {
		return fmt.Errorf("cannot write status line in state %d", w.writerState)
	}
	defer func() { w.writerState = writerStateHeader }()
	_, err := w.writer.Write(getStatusLine(statusCode))
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != writerStateHeader {
		return fmt.Errorf("cannot write headers in state %d", w.writerState)
	}
	defer func() { w.writerState = writerStateBody }()

	for key, value := range headers {
		_, err := w.writer.Write(fmt.Appendf([]byte{}, "%s: %s\r\n", key, value))
		if err != nil {
			return err
		}
	}

	_, err := w.writer.Write([]byte("\r\n"))
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}
	return w.writer.Write(p)
}
