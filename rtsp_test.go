package rtsp

import (
	"bufio"
	"bytes"
	"testing"

	"gotest.tools/v3/assert"
)

func TestResponse(t *testing.T) {
	res, err := NewResponse(StatusNotFound, []byte("Hello world"))
	assert.NilError(t, err)
	res.Header["CSeq"] = "100"
	res.Header["foo"] = "bar"

	// encode it
	var buf bytes.Buffer
	err = res.Write(&buf)
	assert.NilError(t, err)

	// decode it
	res2, err := ReadResponse(bufio.NewReader(&buf))
	assert.NilError(t, err)
	assert.DeepEqual(t, res, res2)
}

func TestRequest(t *testing.T) {
	req, err := NewRequest(MethodOptions, "rtsp://someurl", []byte("what"))
	assert.NilError(t, err)
	req.Header["CSeq"] = "1"
	req.Header["Authorize"] = "secret"

	// encode it
	var buf bytes.Buffer
	err = req.Write(&buf)
	assert.NilError(t, err)

	// decode it
	req2, err := ReadRequest(bufio.NewReader(&buf))
	assert.NilError(t, err)
	assert.DeepEqual(t, req, req2)
}

func TestFrame(t *testing.T) {
	f := Frame{
		Channel: 1,
		Data:    []byte("hello world"),
	}

	// encode it
	var buf bytes.Buffer
	err := f.Write(&buf)
	assert.NilError(t, err)

	// make sure it's recognized as a frame
	br := bufio.NewReader(&buf)
	ok, err := IsFrame(br)
	assert.NilError(t, err)
	assert.Assert(t, ok)

	// decode it
	f2, err := ReadFrame(br)
	assert.NilError(t, err)
	assert.DeepEqual(t, f, f2)
}
