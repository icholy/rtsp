package rtsp

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type Client struct {
	cseq      int
	UserAgent string
	Transport RoundTripper
}

func NewClient() *Client {
	return &Client{
		UserAgent: "Golang-RTSP",
		Transport: &Transport{},
	}
}

func (c *Client) Close() error {
	return c.Transport.Close()
}

func (c *Client) Do(req *Request) (*Response, error) {
	// clone the request so we can modify it
	clone := *req
	clone.Header = cloneHeader(req.Header)
	// add the sequence number
	c.cseq++
	clone.Header.Set("CSeq", strconv.Itoa(c.cseq))
	// add the user-agent
	if c.UserAgent != "" {
		clone.Header.Set("User-Agent", c.UserAgent)
	}
	// make the request
	return c.Transport.RoundTrip(req)
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

func cloneHeader(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}
