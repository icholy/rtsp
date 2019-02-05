package rtsp

type Client struct {
	cseq      int
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

func (c *Client) NextCSeq() int {
	c.cseq++
	return c.cseq
}

func (c *Client) Describe(endpoint string) (*Response, error) {
	req, err := NewRequest(MethodDescribe, endpoint, c.NextCSeq(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/sdp")
	return c.Transport.RoundTrip(req)
}

func (c *Client) Options(endpoint string) (*Response, error) {
	req, err := NewRequest(MethodOptions, endpoint, c.NextCSeq(), nil)
	if err != nil {
		return nil, err
	}
	return c.Transport.RoundTrip(req)
}
