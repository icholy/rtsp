package rtsp

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
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

const (
	// all requests
	StatusContinue = 100

	// all requests
	StatusOK = 200
	// RECORD
	StatusCreated = 201
	// RECORD
	StatusLowOnStorageSpace = 250

	// all requests
	StatusMultipleChoices = 300
	// all requests
	StatusMovedPermanently = 301
	// all requests
	StatusMovedTemporarily = 302
	// all requests
	StatusSeeOther = 303
	// all requests
	StatusUseProxy = 305

	// all requests
	StatusBadRequest = 400
	// all requests
	StatusUnauthorized = 401
	// all requests
	StatusPaymentRequired = 402
	// all requests
	StatusForbidden = 403
	// all requests
	StatusNotFound = 404
	// all requests
	StatusMethodNotAllowed = 405
	// all requests
	StatusNotAcceptable = 406
	// all requests
	StatusProxyAuthenticationRequired = 407
	// all requests
	StatusRequestTimeout = 408
	// all requests
	StatusGone = 410
	// all requests
	StatusLengthRequired = 411
	// DESCRIBE, SETUP
	StatusPreconditionFailed = 412
	// all requests
	StatusRequestEntityTooLarge = 413
	// all requests
	StatusRequestURITooLong = 414
	// all requests
	StatusUnsupportedMediaType = 415
	// SETUP
	StatusInvalidparameter = 451
	// SETUP
	StatusIllegalConferenceIdentifier = 452
	// SETUP
	StatusNotEnoughBandwidth = 453
	// all requests
	StatusSessionNotFound = 454
	// all requests
	StatusMethodNotValidInThisState = 455
	// all requests
	StatusHeaderFieldNotValid = 456
	// PLAY
	StatusInvalidRange = 457
	// SET_PARAMETER
	StatusParameterIsReadOnly = 458
	// all requests
	StatusAggregateOperationNotAllowed = 459
	// all requests
	StatusOnlyAggregateOperationAllowed = 460
	// all requests
	StatusUnsupportedTransport = 461
	// all requests
	StatusDestinationUnreachable = 462

	// all requests
	StatusInternalServerError = 500
	// all requests
	StatusNotImplemented = 501
	// all requests
	StatusBadGateway = 502
	// all requests
	StatusServiceUnavailable = 503
	// all requests
	StatusGatewayTimeout = 504
	// all requests
	StatusRTSPVersionNotSupported = 505
	// all requests
	StatusOptionNotsupport = 551
)

type ResponseWriter interface {
	http.ResponseWriter
}

type Request struct {
	Method        string
	URL           *url.URL
	Proto         string
	ProtoMajor    int
	ProtoMinor    int
	Header        http.Header
	ContentLength int
	Body          io.ReadCloser
}

func (r Request) String() string {
	s := fmt.Sprintf("%s %s %s/%d.%d\r\n", r.Method, r.URL, r.Proto, r.ProtoMajor, r.ProtoMinor)
	for k, v := range r.Header {
		for _, v := range v {
			s += fmt.Sprintf("%s: %s\r\n", k, v)
		}
	}
	s += "\r\n"
	if r.Body != nil {
		str, _ := ioutil.ReadAll(r.Body)
		s += string(str)
	}
	return s
}

func NewRequest(method, urlStr string, cSeq int, body io.ReadCloser) (*Request, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	if u.Port() == "" && u.Scheme == "rtsp" {
		u.Host += ":554"
	}
	req := &Request{
		Method:     method,
		URL:        u,
		Proto:      "RTSP",
		ProtoMajor: 1,
		ProtoMinor: 0,
		Header:     http.Header{},
		Body:       body,
	}
	req.Header.Set("CSeq", strconv.Itoa(cSeq))
	return req, nil
}

type Session struct {
	cSeq    int
	conn    net.Conn
	session string
}

func NewSession() *Session {
	return &Session{}
}

func (s *Session) nextCSeq() int {
	s.cSeq++
	return s.cSeq
}

func (s *Session) Describe(urlStr string) (*Response, error) {
	req, err := NewRequest(MethodDescribe, urlStr, s.nextCSeq(), nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Accept", "application/sdp")

	if s.conn == nil {
		s.conn, err = net.Dial("tcp", req.URL.Host)
		if err != nil {
			return nil, err
		}
	}

	_, err = io.WriteString(s.conn, req.String())
	if err != nil {
		return nil, err
	}
	return ReadResponse(s.conn)
}

func (s *Session) Options(urlStr string) (*Response, error) {
	req, err := NewRequest(MethodOptions, urlStr, s.nextCSeq(), nil)
	if err != nil {
		panic(err)
	}

	if s.conn == nil {
		s.conn, err = net.Dial("tcp", req.URL.Host)
		if err != nil {
			return nil, err
		}
	}

	_, err = io.WriteString(s.conn, req.String())
	if err != nil {
		return nil, err
	}
	return ReadResponse(s.conn)
}

func (s *Session) Setup(urlStr, transport string) (*Response, error) {
	req, err := NewRequest(MethodSetup, urlStr, s.nextCSeq(), nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Transport", transport)

	if s.conn == nil {
		s.conn, err = net.Dial("tcp", req.URL.Host)
		if err != nil {
			return nil, err
		}
	}

	_, err = io.WriteString(s.conn, req.String())
	if err != nil {
		return nil, err
	}
	resp, err := ReadResponse(s.conn)
	s.session = resp.Header.Get("Session")
	return resp, err
}

func (s *Session) Play(urlStr, sessionId string) (*Response, error) {
	req, err := NewRequest(MethodPlay, urlStr, s.nextCSeq(), nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Session", sessionId)

	if s.conn == nil {
		s.conn, err = net.Dial("tcp", req.URL.Host)
		if err != nil {
			return nil, err
		}
	}

	_, err = io.WriteString(s.conn, req.String())
	if err != nil {
		return nil, err
	}
	return ReadResponse(s.conn)
}

type closer struct {
	*bufio.Reader
	r io.Reader
}

func (c closer) Close() error {
	if c.Reader == nil {
		return nil
	}
	defer func() {
		c.Reader = nil
		c.r = nil
	}()
	if r, ok := c.r.(io.ReadCloser); ok {
		return r.Close()
	}
	return nil
}

func ParseRTSPVersion(s string) (proto string, major int, minor int, err error) {
	parts := strings.SplitN(s, "/", 2)
	proto = parts[0]
	parts = strings.SplitN(parts[1], ".", 2)
	if major, err = strconv.Atoi(parts[0]); err != nil {
		return
	}
	if minor, err = strconv.Atoi(parts[0]); err != nil {
		return
	}
	return
}

// super simple RTSP parser; would be nice if net/http would allow more general parsing
func ReadRequest(r io.Reader) (req *Request, err error) {
	req = new(Request)
	req.Header = make(map[string][]string)

	b := bufio.NewReader(r)
	var s string

	// TODO: allow CR, LF, or CRLF
	if s, err = b.ReadString('\n'); err != nil {
		return
	}

	parts := strings.SplitN(s, " ", 3)
	req.Method = parts[0]
	if req.URL, err = url.Parse(parts[1]); err != nil {
		return
	}

	req.Proto, req.ProtoMajor, req.ProtoMinor, err = ParseRTSPVersion(parts[2])
	if err != nil {
		return
	}

	// read headers
	for {
		if s, err = b.ReadString('\n'); err != nil {
			return
		} else if s = strings.TrimRight(s, "\r\n"); s == "" {
			break
		}

		parts := strings.SplitN(s, ":", 2)
		req.Header.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
	}

	req.ContentLength, _ = strconv.Atoi(req.Header.Get("Content-Length"))
	fmt.Println("Content Length:", req.ContentLength)
	req.Body = closer{b, r}
	return
}

type Response struct {
	Proto      string
	ProtoMajor int
	ProtoMinor int

	StatusCode int
	Status     string

	ContentLength int64

	Header http.Header
	Body   io.ReadCloser
}

func (res Response) String() string {
	s := fmt.Sprintf("%s/%d.%d %d %s\n", res.Proto, res.ProtoMajor, res.ProtoMinor, res.StatusCode, res.Status)
	for k, v := range res.Header {
		for _, v := range v {
			s += fmt.Sprintf("%s: %s\n", k, v)
		}
	}
	return s
}

func ReadResponse(r io.Reader) (res *Response, err error) {
	res = new(Response)
	res.Header = make(map[string][]string)

	b := bufio.NewReader(r)
	var s string

	// TODO: allow CR, LF, or CRLF
	if s, err = b.ReadString('\n'); err != nil {
		return
	}

	parts := strings.SplitN(s, " ", 3)
	res.Proto, res.ProtoMajor, res.ProtoMinor, err = ParseRTSPVersion(parts[0])
	if err != nil {
		return
	}

	if res.StatusCode, err = strconv.Atoi(parts[1]); err != nil {
		return
	}

	res.Status = strings.TrimSpace(parts[2])

	// read headers
	for {
		if s, err = b.ReadString('\n'); err != nil {
			return
		} else if s = strings.TrimRight(s, "\r\n"); s == "" {
			break
		}

		parts := strings.SplitN(s, ":", 2)
		res.Header.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
	}

	res.ContentLength, _ = strconv.ParseInt(res.Header.Get("Content-Length"), 10, 64)

	res.Body = closer{b, r}
	return
}
