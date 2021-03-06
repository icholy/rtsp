package rtsp

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"sync"
)

// Client allows sending and recieving rtsp data over a
// socket connection.
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

// NewClient constructs an rtsp Client wrapping a connection.
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

// Option configures a client.
type Option func(*Client)

// WithAuth configures the client authentication.
func WithAuth(a Auth) Option {
	return func(c *Client) { c.auth = a }
}

// WithFrameHandler sets a callback for incoming interleaved
// binary frames.
func WithFrameHandler(handler func(Frame) error) Option {
	return func(c *Client) { c.frameHandler = handler }
}

// WithUserAgent specifies the user-agent to be sent with
// each request.
func WithUserAgent(userAgent string) Option {
	return func(c *Client) { c.userAgent = userAgent }
}

// Do sends a request and reads the response.
func (c *Client) Do(req *Request) (*Response, error) {
	if _, err := c.auth.Authorize(req, nil); err != nil {
		return nil, err
	}
	res, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == StatusUnauthorized {
		retry, err := c.auth.Authorize(req, res)
		if err != nil {
			return nil, err
		}
		if retry {
			return c.roundTrip(req)
		}
	}
	return res, nil
}

// Describe is a helper method for sending an DESCRIBE request.
func (c *Client) Describe(endpoint string) (*Response, error) {
	req, err := NewRequest(MethodDescribe, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/sdp")
	return c.Do(req)
}

// Options is a helper method for sending an OPTIONS request.
func (c *Client) Options(endpoint string) (*Response, error) {
	req, err := NewRequest(MethodOptions, endpoint, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Setup is a helper method for sending a SETUP request.
func (c *Client) Setup(endpoint, transport string) (*Response, error) {
	req, err := NewRequest(MethodSetup, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Transport", transport)
	return c.Do(req)
}

// Session parses the session id from a SETUP response.
func Session(res *Response) (string, error) {
	if err := res.Err(); err != nil {
		return "", err
	}
	session, ok := res.Header.Field("Session", 0)
	if !ok {
		return "", errors.New("missing Sessions header")
	}
	return session, nil
}

// Play is a helper method for sending a PLAY request.
func (c *Client) Play(endpoint, session string) (*Response, error) {
	req, err := NewRequest(MethodPlay, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Session", session)
	req.Header.Set("Range", "npt=0.000-")
	return c.Do(req)
}

// Teardown is a helper method for sending a TEARDOWN request.
func (c *Client) Teardown(endpoint, session string) (*Response, error) {
	req, err := NewRequest(MethodTeardown, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Session", session)
	return c.Do(req)
}

type errResponse struct {
	res *Response
	err error
}

func (c *Client) recv() error {
	ok, err := IsFrame(c.r)
	if err != nil {
		return err
	}
	if ok {
		f, err := ReadFrame(c.r)
		if err != nil {
			return err
		}
		return c.frameHandler(f)
	} else {
		res, err := ReadResponse(c.r)
		c.respCh <- errResponse{res: res, err: err}
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
		return re.res, re.err
	case <-c.doneCh:
		c.errMu.Lock()
		defer c.errMu.Unlock()
		return nil, c.err
	}
}

func (c *Client) roundTrip(req *Request) (*Response, error) {
	// clone the request so we can modify it
	clone := *req
	clone.Header = req.Header.Clone()
	// add the sequence number
	c.cseq++
	clone.Header.Set("CSeq", strconv.Itoa(c.cseq))
	// add the user-agent
	if c.userAgent != "" {
		clone.Header.Set("User-Agent", c.userAgent)
	}
	// make the request
	if err := clone.Write(c.w); err != nil {
		return nil, err
	}
	// wait for a response
	return c.recvResponse()
}
