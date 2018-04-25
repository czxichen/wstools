package command

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

func FileMd5(path string) (str string) {
	File, err := os.Open(path)
	if err != nil {
		return ""
	}
	str = ReadMd5(File)
	File.Close()
	return
}

func ReadMd5(r io.Reader) string {
	h := md5.New()
	io.Copy(h, r)
	return hex.EncodeToString(h.Sum(nil))
}
