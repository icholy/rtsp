package rtsp

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
)

// RTSP Verbs
const (
	MethodAnnounce     = "ANNOUNCE"
	MethodDescribe     = "DESCRIBE"
	MethodGetParameter = "GET_PARAMETER"
	MethodOptions      = "OPTIONS"
	MethodPause        = "PAUSE"
	MethodPlay         = "PLAY"
	MethodRecord       = "RECORD"
	MethodRedirect     = "REDIRECT"
	MethodSetParameter = "SET_PARAMETER"
	MethodSetup        = "SETUP"
	MethodTeardown     = "TEARDOWN"
)

// Request contains all the information required to send a request
type Request struct {
	Method string
	URL    *url.URL
	Proto  string
	Header http.Header
	Body   []byte
}

// Write the request to the provided writer in the wire format.
func (r Request) Write(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "%s %s %s\r\n", r.Method, r.URL, r.Proto); err != nil {
		return err
	}
	if err := r.Header.Write(w); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "\r\n"); err != nil {
		return err
	}
	if _, err := w.Write(r.Body); err != nil {
		return err
	}
	return nil
}

// String returns a string representation of the request.
func (r Request) String() string {
	var s strings.Builder
	if err := r.Write(&s); err != nil {
		return err.Error()
	}
	return s.String()
}

// NewRequest constructs a new request.
// The endpoint must be an absolute url.
// The body may be nil.
func NewRequest(method, endpoint string, body []byte) (*Request, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	req := &Request{
		Method: method,
		URL:    u,
		Proto:  "RTSP/1.0",
		Header: http.Header{},
		Body:   body,
	}
	if len(body) > 0 {
		req.Header.Set("Content-Length", strconv.Itoa(len(body)))
	}
	return req, nil
}

// ReadRequest reads and parses an RTSP request from the provided reader.
func ReadRequest(r *bufio.Reader) (req *Request, err error) {
	tp := textproto.NewReader(r)
	req = new(Request)

	// read response line
	var s string
	if s, err = tp.ReadLine(); err != nil {
		return
	}
	method, url, proto, ok := parseRequestLine(s)
	if !ok {
		return nil, fmt.Errorf("invalid request: %s", s)
	}
	req.Method = method
	req.URL = url
	req.Proto = proto

	// read headers
	header, err := tp.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}
	req.Header = http.Header(header)

	// read body
	if cl := header.Get("Content-Length"); cl != "" {
		length, err := strconv.ParseInt(cl, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid Content-Length: %v", err)
		}
		req.Body = make([]byte, length)
		if _, err := io.ReadFull(r, req.Body); err != nil {
			return nil, err
		}
	}

	return
}

func parseRequestLine(line string) (method string, uri *url.URL, proto string, ok bool) {
	parts := strings.SplitN(line, " ", 3)
	if len(parts) != 3 {
		return
	}
	u, err := url.Parse(parts[1])
	if err != nil {
		return
	}
	return parts[0], u, parts[2], true
}
