package utils

import (
	"bufio"
	"io"
)

func SkipHead(count int, reader *bufio.Reader) {
	for i := 0; i < count; i++ {
		_, _ = reader.ReadString('\n')
	}
}

func ReadLines(count int, ignore int, reader *bufio.Reader) []string {
	lines := make([]string, count-ignore)
	for i := 0; i < count; i++ {
		line, err := reader.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		if i >= ignore {
			lines[i-ignore] = line
		}
	}
	return lines
}
