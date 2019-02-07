package rtsp

type Auth interface {
	// Authorize the request given the response
	// This is called once before the request is send with a nil Response
	// and a second time if the response came back with status code 401
	// unauthorized
	Authorize(*Request, *Response) (bool, error)
}

type noAuth struct{}

func (noAuth) Authorize(req *Request, resp *Response) (bool, error) {
	return false, nil
}
