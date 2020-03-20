package rtsp

import (
	"bufio"
	"strings"
)

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
