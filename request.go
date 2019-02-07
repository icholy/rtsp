package rtsp

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

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

type Request struct {
	Method string
	URL    *url.URL
	Proto  string
	Header http.Header
	Body   []byte
}

func (r Request) WriteTo(w io.Writer) error {
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

func (r Request) String() string {
	var s strings.Builder
	if err := r.WriteTo(&s); err != nil {
		return err.Error()
	}
	return s.String()
}

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
	return req, nil
}
