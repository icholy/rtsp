package rtsp

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
)

type Response struct {
	Proto      string
	ProtoMajor int
	ProtoMinor int

	StatusCode int
	Status     string

	Header http.Header
	Body   []byte
}

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

func (res Response) WriteTo(w io.Writer) error {
	if _, err := fmt.Fprintf(w,
		"%s/%d.%d %d %s\n",
		res.Proto, res.ProtoMajor, res.ProtoMinor, res.StatusCode, res.Status,
	); err != nil {
		return err
	}
	if err := res.Header.Write(w); err != nil {
		return err
	}
	if _, err := w.Write(res.Body); err != nil {
		return err
	}
	return nil
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
	if cl := header.Get("Content-Length"); cl != "" {
		length, err := strconv.ParseInt(cl, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid Content-Length: %v", err)
		}
		res.Body = make([]byte, length)
		if _, err := r.Read(res.Body); err != nil {
			return nil, err
		}
	}

	return
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
