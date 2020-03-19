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

// RTSP response status codes
const (
	StatusContinue                      = 100
	StatusOK                            = 200
	StatusCreated                       = 201
	StatusLowOnStorageSpace             = 250
	StatusMultipleChoices               = 300
	StatusMovedPermanently              = 301
	StatusMovedTemporarily              = 302
	StatusSeeOther                      = 303
	StatusUseProxy                      = 305
	StatusBadRequest                    = 400
	StatusUnauthorized                  = 401
	StatusPaymentRequired               = 402
	StatusForbidden                     = 403
	StatusNotFound                      = 404
	StatusMethodNotAllowed              = 405
	StatusNotAcceptable                 = 406
	StatusProxyAuthenticationRequired   = 407
	StatusRequestTimeout                = 408
	StatusGone                          = 410
	StatusLengthRequired                = 411
	StatusPreconditionFailed            = 412
	StatusRequestEntityTooLarge         = 413
	StatusRequestURITooLong             = 414
	StatusUnsupportedMediaType          = 415
	StatusInvalidparameter              = 451
	StatusIllegalConferenceIdentifier   = 452
	StatusNotEnoughBandwidth            = 453
	StatusSessionNotFound               = 454
	StatusMethodNotValidInThisState     = 455
	StatusHeaderFieldNotValid           = 456
	StatusInvalidRange                  = 457
	StatusParameterIsReadOnly           = 458
	StatusAggregateOperationNotAllowed  = 459
	StatusOnlyAggregateOperationAllowed = 460
	StatusUnsupportedTransport          = 461
	StatusDestinationUnreachable        = 462
	StatusInternalServerError           = 500
	StatusNotImplemented                = 501
	StatusBadGateway                    = 502
	StatusServiceUnavailable            = 503
	StatusGatewayTimeout                = 504
	StatusRTSPVersionNotSupported       = 505
	StatusOptionNotsupport              = 551
)

// Response is a parsed RTSP response.
type Response struct {
	Proto      string
	StatusCode int
	Status     string
	Header     http.Header
	Body       []byte
}

// Write the response to the provided writer in wire format.
func (res Response) Write(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "%s %d %s\n",
		res.Proto, res.StatusCode, res.Status,
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
func NewResponse(code int, status string, body []byte) (*Response, error) {
	res := &Response{
		Proto:      "RTSP/1.0",
		StatusCode: code,
		Status:     status,
		Header:     http.Header{},
		Body:       body,
	}
	if len(body) != 0 {
		res.Header.Set("Content-Length", strconv.Itoa(len(body)))
	}
	return res, nil
}

// ReadResponse reads and parses an RTSP response from the provided reader.
func ReadResponse(r *bufio.Reader) (res *Response, err error) {
	tp := textproto.NewReader(r)
	res = new(Response)

	// read response line
	var s string
	if s, err = tp.ReadLine(); err != nil {
		return
	}
	proto, code, status, ok := parseResponseLine(s)
	if !ok {
		return nil, fmt.Errorf("invalid response: %s", s)
	}
	res.Proto = proto
	res.StatusCode = code
	res.Status = status

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
