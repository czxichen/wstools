package zip

import (
	"archive/tar"
	"compress/gzip"
	"os"
)

type TgzWirter struct {
	tar *tar.Writer
	gz  *gzip.Writer
}

func NewTgzWirter(File *os.File) *TgzWirter {
	gw := gzip.NewWriter(File)
	tw := tar.NewWriter(gw)
	return &TgzWirter{tw, gw}
}

func (self *TgzWirter) Close() error {
	self.tar.Close()
	return self.gz.Close()
}

func (self *TgzWirter) WriteHead(path string, info os.FileInfo) error {
	if path == "." || path == ".." {
		return nil
	}
	head, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	head.Name = path
	if info.IsDir() {
		head.Name += "/"
	}
	return self.tar.WriteHeader(head)
}

func (self *TgzWirter) Write(p []byte) (int, error) {
	return self.tar.Write(p)
}

func (self *TgzWirter) Walk(source string) error {
	return walk(source, self)
}
