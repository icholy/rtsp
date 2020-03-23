package rtsp

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"
)

func TestResponse(t *testing.T) {
	files, err := filepath.Glob("testdata/*.response")
	assert.NilError(t, err)
	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			// decode it
			f, err := os.Open(file)
			assert.NilError(t, err)
			defer f.Close()
			req, err := ReadResponse(bufio.NewReader(f))
			assert.NilError(t, err)
			// encode it
			var buf bytes.Buffer
			err = req.Write(&buf)
			assert.NilError(t, err)
			// decode it again
			req2, err := ReadResponse(bufio.NewReader(&buf))
			assert.NilError(t, err)
			// compare to original
			assert.DeepEqual(t, req, req2)
		})
	}
}

func TestRequest(t *testing.T) {
	files, err := filepath.Glob("testdata/*.request")
	assert.NilError(t, err)
	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			// decode it
			f, err := os.Open(file)
			assert.NilError(t, err)
			defer f.Close()
			req, err := ReadRequest(bufio.NewReader(f))
			assert.NilError(t, err)
			// encode it
			var buf bytes.Buffer
			err = req.Write(&buf)
			assert.NilError(t, err)
			// decode it again
			req2, err := ReadRequest(bufio.NewReader(&buf))
			assert.NilError(t, err)
			// compare to original
			assert.DeepEqual(t, req, req2)
		})
	}
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
