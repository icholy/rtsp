package rtsp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Version is the supported RTSP version
const Version = "RTSP/1.0"

// Response is a parsed RTSP response.
type Response struct {
	StatusCode int
	Status     string
	Header     Header
	Body       []byte
}

// Write the response to the provided writer in wire format.
func (res Response) Write(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "%s %d %s\n",
		Version, res.StatusCode, res.Status,
	); err != nil {
		return err
	}
	if err := res.Header.Write(w); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "\r\n"); err != nil {
		return err
	}
	if _, err := w.Write(res.Body); err != nil {
		return err
	}
	return nil
}

// String returns the string representation of the response.
func (res Response) String() string {
	var s strings.Builder
	if err := res.Write(&s); err != nil {
		return err.Error()
	}
	return s.String()
}

// NewResponse constructs a new response.
// The body may be nil.
func NewResponse(code int, body []byte) (*Response, error) {
	res := &Response{
		StatusCode: code,
		Status:     StatusText(code),
		Header:     Header{},
		Body:       body,
	}
	if len(body) != 0 {
		res.Header["Content-Length"] = strconv.Itoa(len(body))
	}
	return res, nil
}

// ReadResponse reads and parses an RTSP response from the provided reader.
func ReadResponse(r *bufio.Reader) (res *Response, err error) {
	res = new(Response)
	// read response line
	var s string
	if s, err = readLine(r); err != nil {
		return
	}
	proto, code, status, ok := parseResponseLine(s)
	if !ok {
		return nil, fmt.Errorf("invalid response: %s", s)
	}
	if proto != Version {
		return nil, fmt.Errorf("unsuported version: %s", proto)
	}
	res.StatusCode = code
	res.Status = status

	// read headers
	res.Header, err = ReadHeader(r)
	if err != nil {
		return
	}

	// read body
	if cl, ok := res.Header["Content-Length"]; ok {
		length, err := strconv.ParseInt(cl, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid Content-Length: %v", err)
		}
		res.Body = make([]byte, length)
		if _, err := io.ReadFull(r, res.Body); err != nil {
			return nil, err
		}
	}

	return
}

func parseResponseLine(line string) (proto string, code int, status string, ok bool) {
	parts := strings.SplitN(line, " ", 3)
	if len(parts) != 3 {
		return
	}
	code, err := strconv.Atoi(parts[1])
	if err != nil {
		return
	}
	status = strings.TrimSpace(parts[2])
	return parts[0], code, status, true
}
