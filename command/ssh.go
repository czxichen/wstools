package command

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/czxichen/wstools/common/cli"
	"github.com/spf13/cobra"
)

var (
	sshConfig ssh
	// SSH ssh命令
	SSH = &cobra.Command{
		Use: `ssh`,
		Example: `	-C iplist -c ls
	-C iplist -s main.go -d /tmp
	-c ls -u root -p 123456 -H 192.168.1.2:22
	-c ls -u root -P id_rsa -H 192.168.0.129:22
	-u root -p 123456 -H 192.168.1.2:22 -s main.go -d /tmp`,
		Run:   sshRun,
		Short: "使用ssh协议群发命令或发送文件",
		Long: `	通过ssh协议群发命令,每个命令发送都是新的session,当从文件读取主机地址和账户密码的时候,格式为IP:PORT USERNAME PASSWD,使用空白分割,-u -p -H 参数不生效,当发送文件的时候目标的地址可以是目录,当是目录的时候保存的文件名,保存为发送的文件名称.`,
	}
)

type ssh struct {
	config, out  string
	hosts, cmd   string
	sfile, dpath string
	user, passwd string
	privatekey   string
	timeout      int
	hostfile     bool
}

func init() {
	SSH.PersistentFlags().StringVarP(&sshConfig.config, "hosts", "C", "", `从文件读取主机列表和账户密码`)
	SSH.PersistentFlags().StringVarP(&sshConfig.cmd, "cmd", "c", "", `要执行的命令`)
	SSH.PersistentFlags().StringVarP(&sshConfig.hosts, "host", "H", "", `指定Host,多个地址可使用','分割`)
	SSH.PersistentFlags().StringVarP(&sshConfig.sfile, "src", "s", "", `指定要发送文件的路径`)
	SSH.PersistentFlags().StringVarP(&sshConfig.dpath, "dst", "d", "", `指定文件保存路径`)
	SSH.PersistentFlags().StringVarP(&sshConfig.user, "user", "u", "", `指定登录的用户`)
	SSH.PersistentFlags().StringVarP(&sshConfig.passwd, "passwd", "p", "", `指定登录用户密码`)
	SSH.PersistentFlags().StringVarP(&sshConfig.privatekey, "private", "P", "", `使用私钥登录服务器`)
	SSH.PersistentFlags().StringVarP(&sshConfig.out, "out", "o", "", `指定结果输出文件,不指定则直接输出到标准输出`)
	SSH.PersistentFlags().BoolVarP(&sshConfig.hostfile, "hostfile", "f", false, `指定Host从文件读取,指定次参数,-H参数必须是文件路径`)
	SSH.PersistentFlags().IntVarP(&sshConfig.timeout, "timeout", "t", 30, `指定连接超时时间`)
}

func sshRun(cmd *cobra.Command, arg []string) {
	if sshConfig.hosts == "" && sshConfig.config == "" {
		cli.FatalOutput(1, "参数错误,必须指定主机地址或主机配置文件")
	}

	var err error
	var host []string
	var hosts [][]string
	if sshConfig.hosts != "" && !sshConfig.hostfile {
		host = strings.Split(sshConfig.hosts, ",")
	} else {
		if sshConfig.hostfile {
			hosts, err = FileLine(sshConfig.config, 1)
			host = make([]string, 0, len(hosts))
			for _, h := range hosts {
				host = append(host, h[0])
			}
		} else {
			hosts, err = FileLine(sshConfig.config, 3)
		}
		if err != nil {
			cli.FatalOutput(1, "读取主机列表失败:%s\n", err.Error())
		}
	}
	if len(host) <= 0 && len(hosts) <= 0 {
		cli.FatalOutput(1, "主机列表为空\n")
	}

	var output = os.Stdout
	defer output.Close()

	if sshConfig.out != "" {
		output, err = os.Create(sshConfig.out)
		if err != nil {
			cli.FatalOutput(1, "创建结果文件失败:%s\n", err.Error())
		}
	}

	if sshConfig.cmd == "" && (sshConfig.sfile == "" || sshConfig.dpath == "") {
		cli.FatalOutput(1, "参数错误\n")
	}

	var conns = make([]*cli.SSHConnection, 0, len(host))
	var keys []string
	if sshConfig.privatekey != "" {
		keys = append(keys, sshConfig.privatekey)
	}

	if host != nil {
		if sshConfig.user == "" || sshConfig.passwd == "" && sshConfig.privatekey == "" {
			cli.FatalOutput(1, "必须指定用户名,密码或私钥\n")
		}
		for _, h := range host {
			conns = append(conns, &cli.SSHConnection{
				Host: h, User: sshConfig.user, Passwd: sshConfig.passwd, Keys: keys,
			})
		}
	} else {
		for _, info := range hosts {
			conns = append(conns, &cli.SSHConnection{
				Host: info[0], User: info[1], Passwd: info[2], Keys: keys,
			})
		}
	}

	if sshConfig.cmd != "" {
		var retChan = make(chan *cli.SSHResult, 1)
		clients := cli.InitClients(conns, sshConfig.timeout, output)
		if len(clients) != 0 {
			go func() {
				for host, client := range clients {
					cmdChan := make(chan string, 1)
					cmdChan <- sshConfig.cmd
					close(cmdChan)
					if err := cli.SSHSendCommonds(host, client.Client, cmdChan, retChan); err != nil {
						retChan <- &cli.SSHResult{Host: host, Error: err}
					}
				}
			}()
			for i := 0; i < len(clients); i++ {
				ret := <-retChan
				if ret.Error == nil {
					fmt.Fprintf(output, "---------------------------SUCCESS\t%s---------------------------\n%s\n", ret.Host, ret.Data)
				} else {
					fmt.Fprintf(output, "---------------------------FAILD\t%s---------------------------\n%s\n", ret.Host, ret.Error.Error())
				}
			}
			close(retChan)
		}
	} else {
		cli.SSHBatchSendFile(conns, sshConfig.timeout, sshConfig.sfile, sshConfig.dpath, output)
	}
}

// FileLine 按行读取文件
func FileLine(path string, count int) ([][]string, error) {
	File, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer File.Close()
	var list [][]string
	buf := bufio.NewReader(File)
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			if err != io.EOF {
				return list, err
			}
			return list, nil
		}
		l := strings.Fields(string(line))
		if len(l) == count {
			list = append(list, l)
		} else {
			fmt.Printf("无效的数据:%s\n", string(line))
		}
	}
}
