package command

import (
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/lucas-clemente/quic-go/h2quic"
)

func GetIPFromRequest(r *http.Request) string {
	list := strings.Split(r.RemoteAddr, ":")
	if len(list) != 2 {
		return ""
	}
	return list[0]
}

//下载指定url内容保存到本地
func Wget(quic bool, request, save, user, passwd string, tlscfg *tls.Config) error {
	req, err := http.NewRequest("GET", request, nil)
	if err != nil {
		return err
	}
	if user != "" {
		req.SetBasicAuth(user, passwd)
	}
	var client = &http.Client{}
	if strings.HasPrefix(request, "https") || quic {
		if !quic {
			client.Transport = &http.Transport{TLSClientConfig: tlscfg}
		} else {
			client.Transport = &h2quic.RoundTripper{TLSClientConfig: tlscfg}
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	save = filepath.Clean(save)
	info, err := os.Lstat(save)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	if info != nil && info.IsDir() {
		var filename = string(resp.Header.Get("Content-Disposition"))
		save += string(filepath.Separator)
		if len(filename) > 0 {
			for _, name := range strings.Split(filename, ";") {
				if strings.Contains(name, "filename") {
					list := strings.Split(name, "=")
					if len(list) == 2 {
						filename = strings.TrimSpace(list[1])
					}
					break
				}
			}
		}

		if filename == "" {
			list := strings.Split(req.URL.Path, "/")
			filename = list[len(list)-1]
		}
		save += filename
	}
	File, err := os.Create(save)
	if err != nil {
		return err
	}

	defer File.Close()
	_, err = io.Copy(File, resp.Body)
	return err
}
