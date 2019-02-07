package rtsp

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Client struct {
	auth         Auth
	userAgent    string
	frameHandler func(Frame) error

	w    io.Writer
	r    *bufio.Reader
	cseq int

	doneCh chan struct{}
	respCh chan errResponse
	errMu  sync.Mutex
	err    error
}

func NewClient(conn io.ReadWriter, options ...Option) *Client {
	c := &Client{
		w:            conn,
		r:            bufio.NewReader(conn),
		doneCh:       make(chan struct{}),
		respCh:       make(chan errResponse),
		auth:         noAuth{},
		frameHandler: func(Frame) error { return nil },
	}
	for _, o := range options {
		o(c)
	}
	go c.recvLoop()
	return c
}

type Option func(*Client)

func WithAuth(a Auth) Option {
	return func(c *Client) { c.auth = a }
}

func WithFrameHandler(handler func(Frame) error) Option {
	return func(c *Client) { c.frameHandler = handler }
}

func WithUserAgent(userAgent string) Option {
	return func(c *Client) { c.userAgent = userAgent }
}

func (c *Client) Do(req *Request) (*Response, error) {
	if _, err := c.auth.Authorize(req, nil); err != nil {
		return nil, err
	}
	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == StatusUnauthorized {
		retry, err := c.auth.Authorize(req, resp)
		if err != nil {
			return nil, err
		}
		if retry {
			return c.roundTrip(req)
		}
	}
	return resp, nil
}

func (c *Client) Describe(endpoint string) (*Response, error) {
	req, err := NewRequest(MethodDescribe, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/sdp")
	return c.Do(req)
}

func (c *Client) Options(endpoint string) (*Response, error) {
	req, err := NewRequest(MethodOptions, endpoint, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) Setup(endpoint, transport string) (*Response, error) {
	req, err := NewRequest(MethodSetup, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Transport", transport)
	return c.Do(req)
}

func Session(resp *Response) (string, error) {
	if resp.StatusCode != StatusOK {
		return "", errors.New(resp.Status)
	}
	fields := strings.Split(resp.Header.Get("Session"), ";")
	if len(fields) == 0 {
		return "", errors.New("missing Sessions header")
	}
	return strings.TrimSpace(fields[0]), nil
}

func (c *Client) Play(endpoint, session string) (*Response, error) {
	req, err := NewRequest(MethodPlay, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Session", session)
	req.Header.Set("Range", "npt=0.000-")
	return c.Do(req)
}

func (c *Client) Teardown(endpoint, session string) (*Response, error) {
	req, err := NewRequest(MethodTeardown, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Session", session)
	return c.Do(req)
}

type errResponse struct {
	resp *Response
	err  error
}

func (c *Client) recv() error {
	first, err := c.r.Peek(1)
	if err != nil {
		return err
	}
	if first[0] == '$' {
		f, err := ReadFrame(c.r)
		if err != nil {
			return err
		}
		return c.frameHandler(f)
	} else {
		resp, err := ReadResponse(c.r)
		if err != nil {
			log.Println(err)
		}
		c.respCh <- errResponse{resp: resp, err: err}
		return err
	}
}

func (c *Client) recvLoop() {
	for {
		if err := c.recv(); err != nil {
			c.errMu.Lock()
			c.err = err
			c.errMu.Unlock()
			close(c.doneCh)
			return
		}
	}
}

func (c *Client) recvResponse() (*Response, error) {
	select {
	case re := <-c.respCh:
		return re.resp, re.err
	case <-c.doneCh:
		c.errMu.Lock()
		defer c.errMu.Unlock()
		return nil, c.err
	}
}

func (c *Client) roundTrip(req *Request) (*Response, error) {
	// clone the request so we can modify it
	clone := *req
	clone.Header = cloneHeader(req.Header)
	// add the sequence number
	c.cseq++
	clone.Header.Set("CSeq", strconv.Itoa(c.cseq))
	// add the user-agent
	if c.userAgent != "" {
		clone.Header.Set("User-Agent", c.userAgent)
	}
	// make the request
	if err := req.WriteTo(c.w); err != nil {
		return nil, err
	}
	// wait for a response
	return c.recvResponse()
}

func cloneHeader(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}
