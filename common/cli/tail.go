package cli

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strconv"
	"sync"
)

const readbuf = 1 << 10 // 1024

// TailConfig tail config
type TailConfig struct {
	Line   int
	Lines  int
	Size   string
	Offset string
	Output string
	File   string
}

// RunTail 指定tail命令
func RunTail(tail *TailConfig) error {
	File, err := os.OpenFile(tail.File, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	var output io.WriteCloser
	if tail.Output != "" {
		if output, err = os.Create(tail.Output); err != nil {
			File.Close()
			return err
		}
		defer output.Close()
	} else {
		output = os.Stdout
	}
	tFile := &tailFile{file: File, size: 0, offset: 0}
	defer tFile.Close()

	if tail.Offset != "" {
		Offset := parseCompany(tail.Offset)
		Size := parseCompany(tail.Size)
		if Size == 0 && int64(tail.Lines) == 0 {
			return nil
		}
		return tFile.Read(Offset, Size, int64(tail.Lines), output)
	}
	if tail.Line > 0 {
		return tFile.TailLine(tail.Line, tail.Lines, output)
	}
	return nil
}

type tailFile struct {
	mu     sync.Mutex
	file   *os.File
	size   int64
	offset int64
}

func (f *tailFile) read(p []byte) (n int, err error) {
	if f.offset == 0 {
		return 0, io.EOF
	}
	var offset int
	var length = int64(len(p))
	if f.offset >= length {
		f.offset -= length
	} else {
		offset = int(f.offset)
		f.offset = 0
	}
	_, err = f.file.Seek(f.offset, 0)
	if err == nil {
		n, err = f.file.Read(p)
		if offset != 0 && offset < n {
			n = offset
		}
	}
	return
}

func (f *tailFile) Read(offset, size, lines int64, w io.Writer) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.size < offset {
		return io.EOF
	}
	f.file.Seek(offset, 0)
	switch {
	case size > 0:
		if _, err := io.CopyN(w, f.file, size); err != io.EOF {
			return err
		}
	case lines > 0:
		buf := bufio.NewReader(f.file)
		for i := 0; i < int(lines); i++ {
			line, err := buf.ReadBytes('\n')
			if err != nil {
				return err
			}
			if _, err = w.Write(line); err != nil {
				return err
			}
		}
	}
	return nil
}

// ReadLine 暂时用不到
func (f *tailFile) ReadLine() ([]byte, error) {
	var list []*[]byte
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.offset == 0 {
		return nil, io.EOF
	}

	for {
		var buf = make([]byte, 256)
		n, err := f.read(buf[:])
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		index := bytes.LastIndexByte(buf[:n], '\n')
		if index >= 0 {
			f.offset += int64(index)
			f.file.Seek(f.offset, 0)
			buf = buf[index:n]
			list = append(list, &buf)
			break
		} else {
			buf = buf[:n]
			list = append(list, &buf)
		}
	}

	var last int
	var line = make([]byte, len(list)*256)

	for i := len(list) - 1; i >= 0; i-- {
		copy(line[last:], *list[i])
		last += len(*list[i])
	}

	return line[:last], nil
}

// TailLine 从尾部按行读取
func (f *tailFile) TailLine(line, lines int, w io.Writer) (err error) {
	buf := make([]byte, readbuf)
	f.mu.Lock()
	defer f.mu.Unlock()

	var n, l int
	var sep = []byte("\n")

	for {
		if n, err = f.read(buf); err != nil {
			if err == io.EOF {
				f.file.Seek(0, 0)
				break
			}
			return err
		}
		l += bytes.Count(buf[:n], sep)
		if l >= line {
			var seek, i int
			buf = buf[:n]
			for l > line {
				i = bytes.Index(buf, sep) + 1
				buf = buf[i:]
				seek += i
				l--
			}
			f.file.Seek(f.offset+int64(seek), 0)
			break
		}
	}
	if lines == 0 {
		_, err = io.Copy(w, f.file)
	} else {
		buf := bufio.NewReader(f.file)
		var line []byte
		for i := 0; i < lines; i++ {
			if line, err = buf.ReadBytes('\n'); err != nil {
				return
			}
			if _, err = w.Write(line); err != nil {
				return
			}
		}
	}
	return
}

// Close 关闭句柄
func (f *tailFile) Close() error {
	return f.file.Close()
}

func parseCompany(c string) int64 {
	if len(c) < 1 {
		return 0
	}

	if c[len(c)-1] != 'b' {
		c += "b"
	}

	index := len(c) - 2
	var company int64 = 1
	switch c[index] {
	case 'k':
		company = 1 << 10 // 1024
	case 'm':
		company = 1 << 20 // 1024*1024
	default:
		index++
	}
	Size, err := strconv.ParseInt(c[:index], 10, 0)
	if err != nil {
		return -1
	}
	return Size * company
}
