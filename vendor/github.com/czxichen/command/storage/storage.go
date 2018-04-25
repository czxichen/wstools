package storage

import (
	"io"
	"os"
	"sync"
	"sync/atomic"
)

type DiskReadWrite interface {
	Len() int64
	io.ReadWriteCloser
}

type OutPut interface {
	Len() int64
	io.ReadCloser
}

func NewBuffer(body []byte) OutPut {
	return &smallbuffer{eof: 0, size: len(body), body: body}
}

type smallbuffer struct {
	eof  uint32
	size int
	body []byte
	lock sync.Mutex
}

func (self *smallbuffer) Read(p []byte) (int, error) {
	if atomic.LoadUint32(&self.eof) != 0 {
		return 0, io.EOF
	}
	self.lock.Lock()
	defer self.lock.Unlock()

	if len(p) >= len(self.body) {
		copy(p, self.body)
		atomic.SwapUint32(&self.eof, 1)
		return len(self.body), nil
	} else {
		copy(p, self.body[:len(p)])
		self.body = self.body[len(p):]
		return len(p), nil
	}

	return 0, nil
}

func (self *smallbuffer) Len() int64 {
	return int64(self.size)
}

func (self *smallbuffer) Close() error {
	self.body = nil
	return nil
}

//当Job执行的时候,会调用Storage的wirte方法对数据进行持久化
type DiskStorage interface {
	io.WriteCloser
	Sync() error
}

func NewFile(f *os.File) DiskReadWrite {
	return &File{rwc: f}
}

//定义实现Storage接口,用来做windows平台的数据转码存储.
type File struct {
	rwc *os.File
}

func (self *File) Write(p []byte) (int, error) {
	return self.rwc.Write(p)
}

func (self *File) Close() (err error) {
	self.rwc.Sync()
	return self.rwc.Close()
}

func (self *File) Read(p []byte) (int, error) {
	return self.rwc.Read(p)
}

func (self *File) Len() (size int64) {
	f, err := self.rwc.Stat()
	if err != nil {
		size = 0
	} else {
		size = f.Size()
	}
	return
}
