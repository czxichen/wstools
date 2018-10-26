package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type compare struct {
	verbose     bool
	copyto      string
	source      string
	destination string
}

// ComparePath 路径比较
func ComparePath(spath, dpath string, handler func(add bool, src, path string) error) error {
	spath = filepath.Clean(spath)
	dpath = filepath.Clean(dpath)

	spath += string(filepath.Separator)
	dpath += string(filepath.Separator)

	return filepath.Walk(spath, func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		base := strings.TrimPrefix(root, spath)
		match, err := Comparefile(root, dpath+base)
		if err != nil {
			if os.IsNotExist(err) && handler != nil {
				return handler(true, spath, base)
			}
			return err
		}
		if !match && handler != nil {
			return handler(match, spath, base)
		}
		return nil
	})
}

// Comparefile 比较文件
func Comparefile(spath, dpath string) (bool, error) {
	sFile, err := os.Open(spath)
	if err != nil {
		return false, err
	}
	defer sFile.Close()

	dFile, err := os.Open(dpath)
	if err != nil {
		return false, err
	}
	defer dFile.Close()

	sInfo, err := sFile.Stat()
	if err != nil {
		return false, err
	}

	dInfo, err := dFile.Stat()
	if err != nil {
		return false, err
	}

	if dInfo.IsDir() && sInfo.IsDir() {
		return true, nil
	}

	if sInfo.Size() != dInfo.Size() {
		return false, nil
	}

	return Comparebyte(sFile, dFile), nil
}

// Comparebyte 使用字节码比较数据流
func Comparebyte(sfile io.Reader, dfile io.Reader) bool {
	var (
		sint, dint int
		serr, derr error
		sbyte      = make([]byte, 512)
		dbyte      = make([]byte, 512)
	)
	for {
		sint, serr = sfile.Read(sbyte)
		dint, derr = dfile.Read(dbyte)
		if serr != nil || derr != nil {
			if serr == io.EOF && derr == io.EOF {
				return true
			}
			return false
		}
		if sint == dint && bytes.Equal(sbyte, dbyte) {
			continue
		}
		return false
	}
}
