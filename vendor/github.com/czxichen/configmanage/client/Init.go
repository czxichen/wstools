package client

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/czxichen/command/parse"
	"github.com/spf13/cobra"
)

type clientConfig struct {
	CheckMd5    bool   `json:"checkmd5"`
	Home        string `json:"home"`
	RequestMode string `json:"requestmode"`
	MasteUrl    string `json:"masteurl"`
	Primary     string `json:"primary"`
	Action      string `json:"action"`
}

const tmp = "tmp/"

var (
	cfgpath string
	CfgName string
	Config  clientConfig
)

var (
	File         *os.File
	DeployClient = &cobra.Command{
		Use:   "client",
		Short: "快速部署客户端",
		Run:   Client,
	}
)

func init() {
	DeployClient.PersistentFlags().StringVarP(&cfgpath, "config", "c", "", "指定配置文件")
	DeployClient.PersistentFlags().StringVarP(&CfgName, "new", "n", "", "只更新指定服务端配置文件 -n system-cofnig.xml 结合-a getcfg使用")
	DeployClient.PersistentFlags().StringVarP(&Config.Action, "action", "a", "", "指定要做的操作 -a install|getcfg")
	DeployClient.PersistentFlags().StringVarP(&Config.MasteUrl, "master", "m", "127.0.0.1:1789", "指定服务端IP端口")
	DeployClient.PersistentFlags().StringVarP(&Config.RequestMode, "proto", "p", "http", "指定通信协议http|https")
	DeployClient.PersistentFlags().StringVarP(&Config.Home, "home", "H", "./", "指定主目录")
	DeployClient.PersistentFlags().BoolVarP(&Config.CheckMd5, "md5sum", "M", true, "检查主包的md5值")
	DeployClient.PersistentFlags().StringVarP(&Config.Primary, "primary", "P", "", "指定配置关键字")
}

func Client(cmd *cobra.Command, args []string) {
	var err error
	File, err = os.Create("client.log")
	if err != nil {
		log.Fatalln("创建日志文件失败:", err)
	}
	log.SetOutput(File)
	os.MkdirAll(tmp, 0644)
	parseconfig()
	if Config.Action == "" {
		log.Fatalln("必须指定-a参数")
	}
	Config.Home = strings.Replace(Config.Home, "\\", "/", -1)
	if !strings.HasSuffix(Config.Home, "/") {
		Config.Home += "/"
	}

	defer File.Close()
	url := Config.RequestMode + "://" + Config.MasteUrl

	switch Config.Action {
	case "install":
		path := url + "/serverpackage"
		InitPackage(path)
		CreateConfig(CfgName, url, Config.Primary)
	case "getcfg":
		CreateConfig(CfgName, url, Config.Primary)
	default:
		log.Println("-a 参数无效,-h 查看帮助命令.")
	}
}

func parseconfig() {
	if cfgpath == "" {
		return
	}
	buf, err := parse.Parse(cfgpath, "#")
	if err != nil {
		log.Fatalln("读取本地配置文件失败:", err)
	}
	err = json.Unmarshal(buf, &Config)
	if err != nil {
		log.Fatalln("解析本地配置文件失败:", err)
	}
}
