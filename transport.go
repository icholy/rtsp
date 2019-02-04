package rtsp

import (
	"bufio"
	"net"
	"sync"
)

type RoundTripper interface {
	RoundTrip(*Request) (*Response, error)
	Close() error
}

type Transport struct {
	reset  bool
	host   string
	mu     sync.Mutex
	conn   net.Conn
	reader *bufio.Reader
}

func (t *Transport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.closeConn()
}

func (t *Transport) RoundTrip(req *Request) (*Response, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// connect if we need to
	if host := req.URL.Host; host != t.host || t.reset {
		if err := t.closeConn(); err != nil {
			return nil, err
		}
		conn, err := net.Dial("tcp", req.URL.Host)
		if err != nil {
			return nil, err
		}
		t.reset = false
		t.host = host
		t.conn = conn
		t.reader = bufio.NewReader(conn)
	}

	// write request
	if err := req.WriteTo(t.conn); err != nil {
		t.reset = true
		return nil, err
	}
	resp, err := ReadResponse(t.reader)
	if err != nil {
		t.reset = true
		return nil, err
	}

	return resp, err
}

func (t *Transport) closeConn() error {
	if t.conn == nil {
		return nil
	}
	err := t.conn.Close()
	t.conn = nil
	t.host = ""
	return err
}
