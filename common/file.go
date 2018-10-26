package common

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//文件拷贝src源文件,dst目标文件
func Copy(src, dst string) error {
	sFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sFile.Close()
	info, err := sFile.Stat()
	if err != nil {
		return err
	}
	var ret bool = false
retry:
	dFile, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR|os.O_TRUNC, info.Mode())
	if err != nil {
		if os.IsNotExist(err) && !ret {
			if err = os.MkdirAll(filepath.Dir(dst), 0644); err == nil {
				ret = true
				goto retry
			}
		}
		return err
	}
	_, err = io.Copy(dFile, sFile)
	dFile.Close()
	if err == nil {
		err = os.Chtimes(dst, info.ModTime(), info.ModTime())
	}
	return err
}

//拷贝目录文件,目标路径必须是文件夹,如果不存在则会创建,拷贝会把src目录下面的所有文件拷贝到dst目录下
func CopyDir(src, dst string) error {
	src = filepath.FromSlash(src)
	dst = filepath.FromSlash(dst)
	if src[len(src)-1] != filepath.Separator {
		src += string(filepath.Separator)
	}
	if dst[len(dst)-1] != filepath.Separator {
		dst += string(filepath.Separator)
	}
	if err := os.MkdirAll(dst, 0644); err != nil {
		return err
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		ndst := dst + strings.TrimPrefix(path, src)
		if ndst == "" {
			return nil
		}
		if info.IsDir() {
			return os.MkdirAll(ndst, info.Mode())
		}
		return Copy(path, ndst)
	})
}

// FileLine 返回所有行数据,每行必须有count个字段
func FileLine(path string, count int) ([][]string, error) {
	File, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer File.Close()
	var list [][]string
	buf := bufio.NewReader(File)
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			if err != io.EOF {
				return list, err
			}
			return list, nil
		}
		l := strings.Fields(string(line))
		if len(l) == count {
			list = append(list, l)
		}
	}
}
