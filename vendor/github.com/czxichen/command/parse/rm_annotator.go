package parse

import (
	"bufio"
	"bytes"
	"os"
)

func Parse(jsonpath, annotator string) ([]byte, error) {
	var zs = []byte(annotator)
	File, err := os.Open(jsonpath)
	if err != nil {
		return nil, err
	}
	defer File.Close()
	var buf []byte
	b := bufio.NewReader(File)
	for {
		line, _, err := b.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}
		line = bytes.TrimSpace(line)
		if len(line) <= 0 {
			continue
		}
		index := bytes.Index(line, zs)
		if index == 0 {
			continue
		}
		if index > 0 {
			line = line[:index]
		}
		buf = append(buf, line...)
	}
	return buf, nil
}
