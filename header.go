package rtsp

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Header is a collection of key/value pairs belonging to
// a request or response.
type Header map[string]string

// Write the header key/values to the provided writer
func (h Header) Write(w io.Writer) error {
	for key, value := range h {
		line := key + ": " + value + "\r\n"
		if _, err := w.Write([]byte(line)); err != nil {
			return err
		}
	}
	return nil
}

// ReadHeader reads headers from the provided reader
func ReadHeader(r *bufio.Reader) (Header, error) {
	h := Header{}
	for {
		line, err := readLine(r)
		if err != nil {
			return nil, err
		}
		if len(line) == 0 {
			return h, nil
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header: %q", line)
		}
		key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		h[key] = value
	}
}

// Clone returns a copy of the headers
func (h Header) Clone() Header {
	h2 := Header{}
	for k, v := range h {
		h2[k] = v
	}
	return h2
}
