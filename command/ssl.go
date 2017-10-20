package command

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/czxichen/command/rsas"
)

type sslcfg struct {
	exam   bool
	sign   bool
	create bool
	out    string
	rootc  string
	rootk  string
	client string
	config string
}

var (
	sslCFG = sslcfg{}
	SSL    = &Command{
		UsageLine: `ssl -n -c cfg.json -o root`,
		Run:       ssl,
		Short:     "使用rsa对证书简单操作",
		Long: `实现简单的证书生成与签发
	ssl -n -c cfg.json -o agent
	ssl -sign -rc root.crt -rk root.key -ac agent.crt -o sagent	
`,
	}
)

func init() {
	SSL.Flag.BoolVar(&sslCFG.create, "n", false, "-n 创建证书")
	SSL.Flag.BoolVar(&sslCFG.sign, "s", false, "-s 对证书进行签名")
	SSL.Flag.BoolVar(&sslCFG.exam, "e", false, "-e 创建配置样例")
	SSL.Flag.StringVar(&sslCFG.out, "o", "example", "-o agent 输出的名称,自动加后缀")
	SSL.Flag.StringVar(&sslCFG.rootc, "rc", "root.crt", "-rc root.crt 指定根证书路径")
	SSL.Flag.StringVar(&sslCFG.rootk, "rk", "root.key", "-rk root.key 指定根证书路径")
	SSL.Flag.StringVar(&sslCFG.client, "ac", "agent.crt", "-ac agent.crt 指定要签名的证书路径")
	SSL.Flag.StringVar(&sslCFG.config, "c", "", "-c info.json")
}

func ssl(cmd *Command, args []string) bool {
	if sslCFG.exam {
		File, err := os.Create(sslCFG.out + ".json")
		if err != nil {
			log.Fatalf("Create config example error:%s\n", err.Error())
		}
		buf, _ := json.Marshal(rsas.GetDefaultCrtInfo())
		File.Write(buf)
		File.Close()
		return true
	}

	if sslCFG.sign {
		err := rsas.SignerCRTFromFile(sslCFG.rootc, sslCFG.rootk, sslCFG.client, sslCFG.out+".crt")
		if err != nil {
			log.Fatalf("Sign error:%s\n", err.Error())
		}
		return true
	}

	if sslCFG.create {
		buf, err := ioutil.ReadFile(sslCFG.config)
		if err != nil {
			log.Fatalf("Read config error:%s\n", err.Error())
		}

		var info rsas.CertInformation
		err = json.Unmarshal(buf, &info)
		if err != nil {
			log.Fatalf("Unmarshal error:%s\n", err.Error())
		}

		if info.CommonName == "" {
			log.Fatalln("Must specify commonname")
		}

		c, k, err := rsas.CreatePemCRT(nil, nil, info)
		if err != nil {
			log.Fatalf("Create crt error:%s\n", err.Error())
		}
		File, err := os.Create(sslCFG.out + ".crt")
		if err != nil {
			log.Fatalf("Create crt error:%s\n", err.Error())
		}
		File.Write(c)
		File.Close()

		File, err = os.Create(sslCFG.out + ".key")
		if err != nil {
			log.Fatalf("Create key error:%s\n", err.Error())
		}
		File.Write(k)
		File.Close()
		return true
	}
	return false

}
