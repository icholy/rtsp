package rtsp

type Client struct {
	cSeq      int
	session   string
	Transport RoundTripper
}

func NewClient() *Client {
	return &Client{
		Transport: &Transport{},
	}
}

func (c *Client) Close() error {
	return c.Transport.Close()
}

func (c *Client) nextCSeq() int {
	c.cSeq++
	return c.cSeq
}

func (c *Client) Describe(rawurl string) (*Response, error) {
	req, err := NewRequest(MethodDescribe, rawurl, c.nextCSeq(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/sdp")
	return c.Transport.RoundTrip(req)
}

func (c *Client) Options(rawurl string) (*Response, error) {
	req, err := NewRequest(MethodOptions, rawurl, c.nextCSeq(), nil)
	if err != nil {
		return nil, err
	}
	return c.Transport.RoundTrip(req)
}
