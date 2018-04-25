package command

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/czxichen/command/rsas"
	"github.com/spf13/cobra"
)

type RSA_config struct {
	exam   bool
	sign   bool
	create bool
	out    string
	rootc  string
	rootk  string
	client string
	config string
}

var RSA = &cobra.Command{
	Use: `rsa`,
	Example: `	创建证书文件
	-n -c example.json`,
	RunE:  RSA_run,
	Short: "使用rsa对证书简单操作",
	Long:  `实现简单的证书生成与签发`,
}
var _RSA RSA_config

func init() {
	RSA.PersistentFlags().BoolVarP(&_RSA.create, "new", "n", false, "创建证书")
	RSA.PersistentFlags().BoolVarP(&_RSA.sign, "sign", "s", false, "对证书进行签名")
	RSA.PersistentFlags().BoolVarP(&_RSA.exam, "example", "e", false, "创建配置文件")
	RSA.PersistentFlags().StringVarP(&_RSA.out, "out", "o", "example", "输出的名称,自动加后缀")
	RSA.PersistentFlags().StringVar(&_RSA.rootc, "rootcrt", "root.crt", "指定根证书路径")
	RSA.PersistentFlags().StringVar(&_RSA.rootk, "rootkey", "root.key", "指定根证书路径")
	RSA.PersistentFlags().StringVar(&_RSA.client, "agentcrt", "agent.crt", "指定要签名的证书路径")
	RSA.PersistentFlags().StringVarP(&_RSA.config, "config", "c", "", "-c info.json")
}

func RSA_run(cmd *cobra.Command, args []string) error {
	if _RSA.exam {
		File, err := os.Create(_RSA.out + ".json")
		if err != nil {
			fmt.Printf("Create config example error:%s\n", err.Error())
			return nil
		}
		buf, _ := json.Marshal(rsas.GetDefaultCrtInfo())
		File.Write(buf)
		File.Close()
		return nil
	}

	if _RSA.sign {
		err := rsas.SignerCRTFromFile(_RSA.rootc, _RSA.rootk, _RSA.client, _RSA.out+".crt")
		if err != nil {
			fmt.Printf("Sign error:%s\n", err.Error())
		}
		return nil
	}

	if _RSA.create {
		buf, err := ioutil.ReadFile(_RSA.config)
		if err != nil {
			fmt.Printf("Read config error:%s\n", err.Error())
			return nil
		}

		var info rsas.CertInformation
		err = json.Unmarshal(buf, &info)
		if err != nil {
			fmt.Printf("Unmarshal error:%s\n", err.Error())
			return nil
		}

		if info.CommonName == "" {
			fmt.Println("Must specify commonname")
			return nil
		}

		c, k, err := rsas.CreatePemCRT(nil, nil, info)
		if err != nil {
			fmt.Printf("Create crt error:%s\n", err.Error())
			return nil
		}
		File, err := os.Create(_RSA.out + ".crt")
		if err != nil {
			fmt.Printf("Create crt error:%s\n", err.Error())
			return nil
		}
		File.Write(c)
		File.Close()

		File, err = os.Create(_RSA.out + ".key")
		if err != nil {
			fmt.Printf("Create key error:%s\n", err.Error())
			return nil
		}
		File.Write(k)
		File.Close()
		return nil
	}
	return fmt.Errorf("参数错误")
}
