package command

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/czxichen/command/rsas"
	"github.com/czxichen/wstools/common/cli"
	"github.com/spf13/cobra"
)

type rsaConfig struct {
	exam   bool
	sign   bool
	create bool
	out    string
	rootc  string
	rootk  string
	client string
	config string
}

// RSA RSA命令
var RSA = &cobra.Command{
	Use: `rsa`,
	Example: `	创建证书文件
	-n -c example.json`,
	Run:   rsaRun,
	Short: "使用rsa对证书简单操作",
	Long:  `实现简单的证书生成与签发`,
}

var rsaCfg rsaConfig

func init() {
	RSA.PersistentFlags().BoolVarP(&rsaCfg.create, "new", "n", false, "创建证书")
	RSA.PersistentFlags().BoolVarP(&rsaCfg.sign, "sign", "s", false, "对证书进行签名")
	RSA.PersistentFlags().BoolVarP(&rsaCfg.exam, "example", "e", false, "创建配置文件")
	RSA.PersistentFlags().StringVarP(&rsaCfg.out, "out", "o", "example", "输出的名称,自动加后缀")
	RSA.PersistentFlags().StringVar(&rsaCfg.rootc, "rootcrt", "root.crt", "指定根证书路径")
	RSA.PersistentFlags().StringVar(&rsaCfg.rootk, "rootkey", "root.key", "指定根证书路径")
	RSA.PersistentFlags().StringVar(&rsaCfg.client, "agentcrt", "agent.crt", "指定要签名的证书路径")
	RSA.PersistentFlags().StringVarP(&rsaCfg.config, "config", "c", "", "-c info.json")
}

func rsaRun(cmd *cobra.Command, args []string) {
	if rsaCfg.exam {
		File, err := os.Create(rsaCfg.out + ".json")
		if err != nil {
			cli.FatalOutput(1, "Create config example error:%s\n", err.Error())
		}
		buf, _ := json.Marshal(rsas.GetDefaultCrtInfo())
		File.Write(buf)
		File.Close()
		return
	}

	if rsaCfg.sign {
		err := rsas.SignerCRTFromFile(rsaCfg.rootc, rsaCfg.rootk, rsaCfg.client, rsaCfg.out+".crt")
		if err != nil {
			cli.FatalOutput(1, "Sign error:%s\n", err.Error())
		}
		return
	}

	if rsaCfg.create {
		buf, err := ioutil.ReadFile(rsaCfg.config)
		if err != nil {
			cli.FatalOutput(1, "Read config error:%s\n", err.Error())
		}

		var info rsas.CertInformation
		err = json.Unmarshal(buf, &info)
		if err != nil {
			cli.FatalOutput(1, "Unmarshal error:%s\n", err.Error())
		}

		if info.CommonName == "" {
			cli.FatalOutput(1, "Must specify commonname\n")
		}

		c, k, err := rsas.CreatePemCRT(nil, nil, info)
		if err != nil {
			cli.FatalOutput(1, "Create crt error:%s\n", err.Error())
		}
		File, err := os.Create(rsaCfg.out + ".crt")
		if err != nil {
			cli.FatalOutput(1, "Create crt error:%s\n", err.Error())
		}
		File.Write(c)
		File.Close()

		File, err = os.Create(rsaCfg.out + ".key")
		if err != nil {
			cli.FatalOutput(1, "Create key error:%s\n", err.Error())
		}
		File.Write(k)
		File.Close()
	}
}
