package server

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

func getsrvpkg() error {
	F, err := os.Open(Templatedir)
	if err != nil {
		return err
	}
	defer F.Close()

	info, err := F.Readdir(-1)
	if err != nil {
		return err
	}
	var list infos
	for _, i := range info {
		if strings.Contains(i.Name(), "server_") && strings.Contains(i.Name(), ".zip") {
			list = append(list, i)
		}
	}
	if len(list) <= 0 {
		return errors.New("Not found valid server package")
	}
	sort.Sort(list)
	spkg := list[0].Name()
	str := strings.TrimLeft(spkg, "server_")
	str = strings.TrimRight(str, ".zip")
	m, err := getmd5(Templatedir + spkg)
	if err != nil {
		return err
	}
	if strings.ToLower(str) != m {
		log.Printf("'%s'\n'%s'\n", strings.ToLower(str), m)
		return errors.New(spkg + " md5 unmatch")
	}
	serverpkg = spkg
	return nil
}

func getmd5(path string) (string, error) {
	File, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer File.Close()
	m := md5.New()
	_, err = io.Copy(m, File)
	if err != nil {
		log.Println(err)
	}
	return fmt.Sprintf("%x", string(m.Sum([]byte{}))), nil
}

func getpath(key string) map[string][]string {
	return Pathrelation
}

func getvalue(key string) (m map[string]string, ok bool) {
	var values []string
	values, ok = Variables[key]
	if !ok {
		for _, values = range Variables {
			if ok {
				break
			}
			for _, v := range values {
				if v == key {
					ok = true
					break
				}
			}
		}
		if !ok {
			return
		}
	}
	head := Variables["_relationVariable_"]
	m = make(map[string]string)
	for index, k := range head {
		m[k] = values[index]
	}
	return
}

type infos []os.FileInfo

func (self infos) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self infos) Less(i, j int) bool {
	return self[i].ModTime().Unix() > self[j].ModTime().Unix()
}

func (self infos) Len() int {
	return len(self)
}

func (self infos) Sort() {
	sort.Sort(self)
}
