package auth

import (
	"encoding/base64"

	"github.com/beatgammit/rtsp"
)

type BasicAuth struct {
	Username string
	Password string
}

func Basic(username, password string) BasicAuth {
	return BasicAuth{username, password}
}

func (a BasicAuth) Authorize(req *rtsp.Request, resp *rtsp.Response) (bool, error) {
	if resp != nil {
		return true, nil
	}
	req.Header.Set("Authorization", "Basic "+a.encoded())
	return true, nil
}

func (a BasicAuth) encoded() string {
	auth := a.Username + ":" + a.Password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
