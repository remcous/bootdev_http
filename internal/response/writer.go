package response

import (
	"fmt"
	"io"
	"strings"

	"github.com/remcous/bootdev_http/internal/headers"
)

type Writer struct {
	writer      io.Writer
	writerState writerState

	chunkSize int
}

const defaultChunkSize = 1024

type writerState int

const (
	writerStateStatusLine writerState = iota
	writerStateHeader
	writerStateBody
	writerStateTrailers
)

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer:      w,
		writerState: writerStateStatusLine,
		chunkSize:   defaultChunkSize,
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

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}
	chunkSize := len(p)

	totalBytesWriten := 0

	n, err := fmt.Fprintf(w.writer, "%x\r\n", chunkSize)
	if err != nil {
		return totalBytesWriten, err
	}
	totalBytesWriten += n

	n, err = w.writer.Write(p)
	if err != nil {
		return totalBytesWriten, err
	}
	totalBytesWriten += n

	n, err = w.writer.Write([]byte("\r\n"))
	if err != nil {
		return totalBytesWriten, err
	}
	totalBytesWriten += n

	return totalBytesWriten, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	defer func() { w.writerState = writerStateTrailers }()
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}
	return w.writer.Write([]byte("0\r\n"))
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.writerState != writerStateTrailers {
		return fmt.Errorf("cannot write trailers in state %d", w.writerState)
	}

	trailersRaw, exists := h.Get("Trailer")
	if !exists {
		w.writer.Write([]byte("\r\n"))
		return nil
	}

	trailers := strings.Split(trailersRaw, ", ")

	for _, trailer := range trailers {
		val, exists := h.Get(trailer)
		if !exists {
			continue
		}

		w.writer.Write(fmt.Appendf([]byte{}, "%s: %s\r\n", trailer, val))
	}

	w.writer.Write([]byte("\r\n"))
	return nil
}
