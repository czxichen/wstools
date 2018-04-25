package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	Cfg          Config
	cpath        string
	serverpkg    string
	Templatedir  string
	Variables    map[string][]string
	Pathrelation map[string][]string
)

var DeployServer = &cobra.Command{
	Use:   "server",
	Short: "快速部署服务端",
	RunE:  Server,
}

func init() {
	DeployServer.PersistentFlags().StringVarP(&cpath, "config", "C", "", "从文件读取配置文件")
	DeployServer.PersistentFlags().StringVarP(&Cfg.IP, "listen", "l", ":1789", "指定监听的地址端口")
	DeployServer.PersistentFlags().StringVarP(&Cfg.Proto, "proto", "p", "http", "指定通信协议,http|https")
	DeployServer.PersistentFlags().StringVarP(&Cfg.CrtPath, "crt", "c", "", "指定证书crt文件")
	DeployServer.PersistentFlags().StringVarP(&Cfg.Keypath, "key", "k", "", "指定证书key文件")
	DeployServer.PersistentFlags().StringVarP(&Cfg.Logname, "log", "L", "server.log", "指定日志路径")
	DeployServer.PersistentFlags().StringVarP(&Cfg.Download, "download", "d", "./", "指定下载文件所在的目录")
}

func readconfig(cfgpath string, cfg *Config) {
	File, err := os.Open(cfgpath)
	if err != nil {
		log.Fatalln(err)
	}
	defer File.Close()
	var buf []byte
	b := bufio.NewReader(File)
	for {
		line, _, err := b.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Fatalln(err)
		}
		line = bytes.TrimSpace(line)
		if len(line) <= 0 {
			continue
		}
		index := bytes.Index(line, []byte("#"))
		if index == 0 {
			continue
		}
		if index > 0 {
			line = line[:index]
		}
		buf = append(buf, line...)
	}
	err = json.Unmarshal(buf, &cfg)
	if err != nil {
		log.Println(string(buf))
		log.Fatalln("解析配置文件失败:", err)
	}
}

type serverInfo struct {
	Variable map[string]string
	Relation map[string][]string
}

type Config struct {
	IP       string `json:"ip"`
	Proto    string `json:"proto"`
	CrtPath  string `json:"crtpath"`
	Keypath  string `json:"keypath"`
	Logname  string `json:"logname"`
	Download string `json:"download"`
}

func Parseconfig() {
	log.Println("开始读取变量和模版文件")
	var err error
	Pathrelation, Variables, err = relationConfig(Templatedir + "server.xlsx")
	if err != nil {
		log.Fatalln("读取关系表失败:", err)
	}

	log.Printf("server.xlsx解析成功\n")
	log.Printf("路径关系:\n")
	for key, value := range Pathrelation {
		log.Printf("文件名:%s\t路径:%v\n", key, value)
	}
	log.Printf("变量:\n")
	for key, value := range Variables {
		log.Printf("索引:%s\t值:%v\n", key, value)
	}
	var list []string
	for k, _ := range Pathrelation {
		list = append(list, k)
	}
	log.Println("开始打包模版文件")
	err = Zip(Cfg.Download, list)
	if err != nil {
		log.Fatalln("打包模版文件失败:", err)
	}
	log.Println("开始获取最新的服务端包")
	err = getsrvpkg()
	if err != nil {
		log.Fatalln("检查server包出错:", err)
	}
	log.Printf("服务端包名称:%s\n", serverpkg)
}
