package auth

import (
	"net/http"

	"github.com/icholy/digest"
	"github.com/icholy/rtsp"
)

// Digest is an Auth implementation for the digest authentication.
type Digest struct {
	Username string
	Password string
}

// WithDigest returns a client option for using digest auth
func WithDigest(username, password string) rtsp.Option {
	return rtsp.WithAuth(Digest{
		Username: username,
		Password: password,
	})
}

// Authorize the request.
func (a Digest) Authorize(req *rtsp.Request, resp *rtsp.Response) (bool, error) {
	if resp == nil {
		return true, nil
	}
	chal, err := digest.FindChallenge(http.Header(resp.Header))
	if err != nil {
		return false, err
	}
	cred, err := digest.Digest(chal, digest.Options{
		Method:   req.Method,
		URI:      req.URL.RequestURI(),
		Username: a.Username,
		Password: a.Password,
	})
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", cred.String())
	return true, nil
}
