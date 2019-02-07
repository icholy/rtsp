package rtsp

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	// Client to server for presentation and stream objects; recommended
	MethodDescribe = "DESCRIBE"
	// Bidirectional for client and stream objects; optional
	MethodAnnounce = "ANNOUNCE"
	// Bidirectional for client and stream objects; optional
	MethodGetParameter = "GET_PARAMETER"
	// Bidirectional for client and stream objects; required for Client to server, optional for server to client
	MethodOptions = "OPTIONS"
	// Client to server for presentation and stream objects; recommended
	MethodPause = "PAUSE"
	// Client to server for presentation and stream objects; required
	MethodPlay = "PLAY"
	// Client to server for presentation and stream objects; optional
	MethodRecord = "RECORD"
	// Server to client for presentation and stream objects; optional
	MethodRedirect = "REDIRECT"
	// Client to server for stream objects; required
	MethodSetup = "SETUP"
	// Bidirectional for presentation and stream objects; optional
	MethodSetParameter = "SET_PARAMETER"
	// Client to server for presentation and stream objects; required
	MethodTeardown = "TEARDOWN"
)

type Request struct {
	Method string
	URL    *url.URL
	Proto  string
	Header http.Header
	Body   []byte
}

func (r Request) WriteTo(w io.Writer) error {
	if _, err := fmt.Fprintf(w,
		"%s %s %s\r\n",
		r.Method, r.URL, r.Proto,
	); err != nil {
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
