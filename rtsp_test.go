package rtsp

import (
	"bufio"
	"bytes"
	"testing"

	"gotest.tools/v3/assert"
)

func TestResponse(t *testing.T) {
	res, err := NewResponse(StatusNotFound, "Not Found", []byte("Hello world"))
	assert.NilError(t, err)

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

	// decode it
	f2, err := ReadFrame(&buf)
	assert.NilError(t, err)
	assert.DeepEqual(t, f, f2)
}
