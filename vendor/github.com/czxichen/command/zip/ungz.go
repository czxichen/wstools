package zip

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"strings"
)

func Ungzip(filepath, desdir string, Log func(format string, v ...interface{})) error {
	File, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer File.Close()
	desdir = strings.Replace(desdir, "\\", "/", -1)
	if !strings.HasSuffix(desdir, "/") {
		desdir = desdir + "/"
	}
	gw, err := gzip.NewReader(File)
	if err != nil {
		return err
	}
	defer gw.Close()
	tw := tar.NewReader(gw)
	for {
		head, err := tw.Next()
		if err != nil {
			if err.Error() == "EOF" {
				return nil
			}
			return err
		}
		if head.FileInfo().IsDir() {
			err := os.MkdirAll(desdir+head.Name, os.FileMode(head.Mode))
			if err != nil {
				return err
			}
			if Log != nil {
				Log("create directory: %s\n", desdir+head.Name)
			}
			continue
		}
		F, err := os.OpenFile(desdir+head.Name, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(head.Mode))
		if err != nil {
			continue
		}
		io.Copy(F, tw)
		F.Close()
		os.Chtimes(desdir+head.Name, head.AccessTime, head.ModTime)
		if Log != nil {
			Log("create file: %s\n", desdir+head.Name)
		}
	}
	return nil
}
