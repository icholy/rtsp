package rtsp

import (
	"bufio"
	"fmt"
	"strings"
)

const version = "RTSP/1.0"

func checkVersion(v string) error {
	if v != version {
		return fmt.Errorf("unsuported version: %q", v)
	}
	return nil
}

func readLine(r *bufio.Reader) (string, error) {
	var line strings.Builder
	for {
		s, more, err := r.ReadLine()
		if err != nil {
			return "", err
		}
		line.Write(s)
		if !more {
			return line.String(), nil
		}
	}
}
