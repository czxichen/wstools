package zip

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Compress interface {
	Close() error
	Walk(source string) error
	WriteHead(path string, info os.FileInfo) error
	Write(p []byte) (int, error)
}

//path=/path/dir 则打包的时候会加dir目录,如果path=/path/dir/则不打包dir目录
func walk(path string, compresser Compress) error {
	var (
		opath     string = filepath.FromSlash(path)
		baseDir   string = ""
		separator string = string(filepath.Separator)
	)

	path = filepath.Clean(path) + separator

	if !strings.HasSuffix(opath, separator) {
		baseDir = filepath.Base(path) + separator
	}

	return filepath.Walk(path, func(baseroot string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		var root = strings.TrimPrefix(baseroot, path)
		if root == "" {
			return nil
		}
		err = compresser.WriteHead(baseDir+root, info)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		File, err := os.Open(baseroot)
		if err != nil {
			return err
		}
		_, err = io.Copy(compresser, File)
		File.Close()
		return err
	})
}
