package auth

import (
	"encoding/base64"

	"github.com/beatgammit/rtsp"
)

// Basic is an Auth implementation for basic authentication.
type Basic struct {
	Username string
	Password string
}

// Authorize the request
func (a Basic) Authorize(req *rtsp.Request, resp *rtsp.Response) (bool, error) {
	if resp != nil {
		return true, nil
	}
	req.Header.Set("Authorization", "Basic "+a.encoded())
	return true, nil
}

func (a Basic) encoded() string {
	auth := a.Username + ":" + a.Password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
