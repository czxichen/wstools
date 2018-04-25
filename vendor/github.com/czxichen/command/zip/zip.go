package zip

import (
	"archive/zip"
	"io"
	"os"
	"time"
)

func NewZipWriter(File *os.File) *ZipWrite {
	zipwrite := zip.NewWriter(File)
	return &ZipWrite{zone: 8, zw: zipwrite, file: File}
}

type ZipWrite struct {
	zone   int64
	zw     *zip.Writer
	writer io.Writer
	file   *os.File
}

func (self *ZipWrite) Close() error {
	self.zw.Close()
	return self.file.Close()
}

func (self *ZipWrite) WriteHead(path string, info os.FileInfo) error {
	if path == "." || path == ".." {
		return nil
	}
	head, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		head.Method = zip.Deflate
	}
	head.Name = path
	if info.IsDir() {
		head.Name += "/"
	}
	head.SetModTime(time.Unix(info.ModTime().Unix()+self.zone*60*60, 0))
	write, err := self.zw.CreateHeader(head)
	if err != nil {
		return err
	}
	self.writer = write
	return nil
}

func (self *ZipWrite) Write(p []byte) (int, error) {
	return self.writer.Write(p)
}

func (self *ZipWrite) Walk(source string) error {
	return walk(source, self)
}
