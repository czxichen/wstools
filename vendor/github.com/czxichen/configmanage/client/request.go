package client

import (
	"encoding/gob"
	"errors"
	"io"
	"net/http"
	"os"
)

func getvalues(url string) (serverInfo, error) {
	m := serverInfo{}
	resp, err := get(url)
	if err != nil {
		return m, err
	}
	defer resp.Body.Close()
	err = gob.NewDecoder(resp.Body).Decode(&m)
	if err != nil {
		return m, err
	}
	return m, nil
}

func download(url string) (string, error) {
	resp, err := get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	filename := resp.Header.Get("filename")
	if filename == "" {
		return "", errors.New("not a valid file")
	}
	File, err := os.Create(tmp + filename)
	if err != nil {
		return "", err
	}
	defer File.Close()
	io.Copy(File, resp.Body)
	return filename, nil
}

func get(url string) (resp *http.Response, err error) {
	resp, err = http.Get(url)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		err = errors.New("404 错误")
		return
	}
	return
}
