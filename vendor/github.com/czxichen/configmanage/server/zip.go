package server

import (
	"io"
	"log"
	"os"

	"github.com/czxichen/command/zip"
)

func Zip(dir string, list []string) error {
	dirinfo, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	var File *os.File
	if _, err = os.Lstat(dir + "configtemp.zip"); err == nil {
		File, err = os.Create(dir + "configtemp.zip.tmp")
		defer func() {
			for i := 0; i < 5; i++ {
				err := os.Remove(dir + "configtemp.zip")
				if err != nil {
					if i == 4 {
						log.Println("更新configtemp出错:", err)
						os.Exit(1)
					}
					continue
				}
				break
			}
			os.Rename(dir+"configtemp.zip.tmp", dir+"configtemp.zip")
		}()
	} else {
		File, err = os.Create(dir + "configtemp.zip")
	}
	if err != nil {
		return err
	}
	z := zip.NewZipWriter(File)
	defer z.Close()
	z.WriteHead("configtemp", dirinfo)
	for _, path := range list {
		f, err := os.Open(dir + "template/" + path)
		if err != nil {
			return err
		}
		info, err := f.Stat()
		if err != nil {
			return err
		}
		err = z.WriteHead("configtemp/"+path, info)
		if err != nil {
			f.Close()
			return err
		}
		io.Copy(z, f)
		f.Close()
	}
	return nil
}
