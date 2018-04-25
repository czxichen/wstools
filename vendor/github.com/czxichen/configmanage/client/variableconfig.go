package client

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/czxichen/command/zip"
)

type serverInfo struct {
	Variable map[string]string
	Relation map[string][]string
}

var info serverInfo

//url = "http://127.0.0.1:1789/getvalues?key=7400046"
func initInfo(url string) (err error) {
	info, err = getvalues(url)
	return
}

//url="http://127.0.0.1:1789/serverpackage"
func InitPackage(url string) {
	filename, err := download(url)
	if err != nil {
		log.Fatalf("下载服务端包出错:%s\n", err.Error())
	}
	if getmd5(filename) {
		zip.Unzip(tmp+filename, Config.Home, log.Printf)
	} else {
		log.Fatalf("%s md5校验失败\n", filename)
	}
}

//url = "http://127.0.0.1:1789/configtemplate"
func initConfig(url string) error {
	filename, err := download(url)
	if err != nil {
		return err
	}
	err = zip.Unzip(tmp+filename, tmp, log.Printf)
	if err != nil {
		return err
	}
	return nil
}

var temp *template.Template

func initTemplate(host, key string) {
	var url string = host + "/getvalues"
	if len(key) > 0 {
		url += "?key=" + key
	}
	err := initInfo(url)
	if err != nil {
		log.Println("获取变量失败:", err)
		os.Exit(1)
	}
	url = host + "/configtemplate"
	err = initConfig(url)
	if err != nil {
		log.Fatalf("初始化配置文件失败:%s\n", err.Error())
	}
	var list []string
	for key, _ := range info.Relation {
		list = append(list, tmp+"configtemp/"+key)
	}
	temp, err = template.ParseFiles(list...)
	if err != nil {
		log.Fatalf("加载模版文件失败:%s\n", err.Error())
	}
}

func CreateConfig(filename, host, key string) {
	initTemplate(host, key)
	if filename != "" {
		paths, ok := info.Relation[filename]
		if !ok {
			log.Printf("没有找到对应的路径配置:%s\n", filename)
			return
		}
		create(paths, filename)
		return
	}

	for name, paths := range info.Relation {
		create(paths, name)
	}
}

func create(paths []string, name string) {
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			path = Config.Home + path
		}
		log.Printf("开始创建:%s\n", path)
		File, err := os.Create(path)
		if err != nil {
			log.Printf("创建配置文件:%s失败,错误信息:%s\n", path, err)
			continue
		}
		temp.ExecuteTemplate(File, name, info.Variable)
		File.Close()
	}
}

func getmd5(path string) bool {
	File, err := os.Open("tmp/" + path)
	if err != nil {
		return false
	}
	defer File.Close()
	m := md5.New()
	io.Copy(m, File)
	M := fmt.Sprintf("%x", string(m.Sum([]byte{})))

	str := strings.TrimLeft(path, "server_")
	str = strings.TrimRight(str, ".zip")

	if strings.ToLower(str) != M {
		log.Printf("标注md5:'%s'\n计算md5:'%s'\n", strings.ToLower(str), M)
		return false
	}
	return true
}
