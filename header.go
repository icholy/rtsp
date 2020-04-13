package rtsp

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Header is a collection of key/value pairs belonging to
// a request or response.
type Header map[string][]string

// Set sets the header value
func (h Header) Set(name, value string) {
	h[name] = []string{value}
}

// Add adds the header value
func (h Header) Add(name, value string) {
	h[name] = append(h[name], value)
}

// Get returns the first header value with the provided name.
// If not found, an empty string is returned.
func (h Header) Get(name string) string {
	if len(h[name]) == 0 {
		return ""
	}
	return h[name][0]
}

// Param looks up the header by name and returns the corresponding value for
// the provided key. The expected format is key1=value1;key2=value2 ...
func (h Header) Param(name, key string) (string, bool) {
	for _, p := range strings.Split(h.Get(name), ";") {
		param := strings.SplitN(p, "=", 2)
		if len(param) == 2 && strings.TrimSpace(param[0]) == key {
			return strings.TrimSpace(param[1]), true
		}
	}
	return "", false
}

// Field looks up the header by name and returns the field for the provided
// index. The expected format is field1;field2;field3 ...
func (h Header) Field(name string, index int) (string, bool) {
	fields := strings.Split(h.Get(name), ";")
	if index < 0 || index >= len(fields) {
		return "", false
	}
	return strings.TrimSpace(fields[index]), true
}

// Write the header key/values to the provided writer
func (h Header) Write(w io.Writer) error {
	for key, values := range h {
		for _, value := range values {
			line := key + ": " + value + "\r\n"
			if _, err := w.Write([]byte(line)); err != nil {
				return err
			}
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
		index := strings.IndexByte(line, ':')
		if index == -1 {
			return nil, fmt.Errorf("invalid header: %q", line)
		}
		key := strings.TrimSpace(line[:index])
		h.Add(key, strings.TrimSpace(line[index+1:]))
	}
}

// Clone returns a copy of the headers
func (h Header) Clone() Header {
	h2 := make(Header, len(h))
	for k, v := range h {
		v2 := make([]string, len(v))
		copy(v2, v)
		h2[k] = v2
	}
	return h2
}
