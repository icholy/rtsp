package rtsp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/textproto"
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
	ContentLength int64
	Body          io.ReadCloser
}

func (r Request) WriteTo(w io.Writer) error {
	if _, err := fmt.Fprintf(w,
		"%s %s %s/%d.%d\r\n",
		r.Method, r.URL, r.Proto, r.ProtoMajor, r.ProtoMinor,
	); err != nil {
		return err
	}
	if err := r.Header.Write(w); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "\r\n"); err != nil {
		return err
	}
	if r.Body != nil {
		if _, err := io.Copy(w, r.Body); err != nil {
			return err
		}
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

func NewRequest(method, rawurl string, cSeq int, body io.ReadCloser) (*Request, error) {
	u, err := url.Parse(rawurl)
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
	cSeq      int
	session   string
	Transport RoundTripper
}

func NewSession() *Session {
	return &Session{
		Transport: &Transport{},
	}
}

func (s *Session) Close() error {
	return s.Transport.Close()
}

func (s *Session) nextCSeq() int {
	s.cSeq++
	return s.cSeq
}

func (s *Session) Describe(rawurl string) (*Response, error) {
	req, err := NewRequest(MethodDescribe, rawurl, s.nextCSeq(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/sdp")
	return s.Transport.RoundTrip(req)
}

func (s *Session) Options(rawurl string) (*Response, error) {
	req, err := NewRequest(MethodOptions, rawurl, s.nextCSeq(), nil)
	if err != nil {
		return nil, err
	}
	return s.Transport.RoundTrip(req)
}

func (s *Session) Setup(rawurl, transport string) (*Response, error) {
	req, err := NewRequest(MethodSetup, rawurl, s.nextCSeq(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Transport", transport)
	resp, err := s.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	s.session = resp.Header.Get("Session")
	return resp, nil
}

func (s *Session) Play(rawurl, sessionID string) (*Response, error) {
	req, err := NewRequest(MethodPlay, rawurl, s.nextCSeq(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Session", sessionID)
	return s.Transport.RoundTrip(req)
}

func ParseRTSPVersion(s string) (proto string, major int, minor int, err error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("invalid proto: %s", s)
		return
	}
	proto = parts[0]
	parts = strings.SplitN(parts[1], ".", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("invalid proto: %s", s)
		return
	}
	if major, err = strconv.Atoi(parts[0]); err != nil {
		return
	}
	if minor, err = strconv.Atoi(parts[0]); err != nil {
		return
	}
	return
}

// super simple RTSP parser; would be nice if net/http would allow more general parsing
func ReadRequest(r *bufio.Reader) (req *Request, err error) {
	tp := textproto.NewReader(r)
	req = new(Request)

	// read request line
	var s string
	if s, err = tp.ReadLine(); err != nil {
		return nil, err
	}
	parts := strings.SplitN(s, " ", 3)
	req.Method = parts[0]
	if req.URL, err = url.Parse(parts[1]); err != nil {
		return nil, err
	}
	req.Proto, req.ProtoMajor, req.ProtoMinor, err = ParseRTSPVersion(parts[2])
	if err != nil {
		return nil, err
	}

	// read headers
	header, err := tp.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}
	req.Header = http.Header(header)

	// read body
	req.ContentLength, req.Body, err = readBody(req.Header, r)
	if err != nil {
		return nil, err
	}
	return req, nil
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

func (res Response) WriteTo(w io.Writer) error {
	if _, err := fmt.Fprintf(w,
		"%s/%d.%d %d %s\n",
		res.Proto, res.ProtoMajor, res.ProtoMinor, res.StatusCode, res.Status,
	); err != nil {
		return err
	}
	return res.Header.Write(w)
}

func (res Response) String() string {
	var s strings.Builder
	if err := res.WriteTo(&s); err != nil {
		return err.Error()
	}
	return s.String()
}

func ReadResponse(r *bufio.Reader) (res *Response, err error) {
	tp := textproto.NewReader(r)
	res = new(Response)

	// read response line
	var s string
	if s, err = tp.ReadLine(); err != nil {
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
	header, err := tp.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}
	res.Header = http.Header(header)

	// read body
	res.ContentLength, res.Body, err = readBody(res.Header, r)
	if err != nil {
		return nil, err
	}

	return
}

func readBody(h http.Header, r *bufio.Reader) (int64, io.ReadCloser, error) {
	if cl := h.Get("Content-Length"); cl != "" {
		length, err := strconv.ParseInt(cl, 10, 64)
		if err != nil {
			return 0, nil, fmt.Errorf("invalid Content-Length: %v", err)
		}
		body := make([]byte, length)
		if _, err := r.Read(body); err != nil {
			return 0, nil, err
		}
		return length, ioutil.NopCloser(bytes.NewReader(body)), nil
	}
	return -1, http.NoBody, nil
}
