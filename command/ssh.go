package command

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
)

var (
	lgInfoCFG lgInfo
	SSH       = &Command{
		UsageLine: `ssh -c ls -u root -p 123456 -h 192.168.1.2:22`,
		Run:       lginfo,
		Short:     "使用ssh协议群发命令或发送文件",
		Long: `通过ssh协议群发命令,每个命令发送都是新的session,当从文件读取主机地址和账户密码的时候,格式
为IP:PORT USERNAME PASSWD,使用空白分割,-u -p -H 参数不生效,当发送文件的时候目标的地址可以是目录,当是目
录的时候保存的文件名,保存为发送的文件名称.
	wstools ssh -C iplist -c ls
	wstools ssh -C iplist -s main.go -d /tmp
	wstools ssh -c ls -u root -p 123456 -H 192.168.1.2:22
	wstools ssh -u root -p 123456 -H 192.168.1.2:22 -s main.go -d /tmp
`,
	}
)

type lgInfo struct {
	config, out  string
	hosts, cmd   string
	sfile, dpath string
	user, passwd string
}

func init() {
	SSH.Flag.StringVar(&lgInfoCFG.config, "C", "", `-C="iplist.txt" 从文件读取主机列表和账户密码`)
	SSH.Flag.StringVar(&lgInfoCFG.cmd, "c", "", `-c="ls -a" 要执行的命令`)
	SSH.Flag.StringVar(&lgInfoCFG.hosts, "H", "", `-H="192.164.1.2:22" 指定Host,多个地址可使用','分割`)
	SSH.Flag.StringVar(&lgInfoCFG.sfile, "s", "", `-s=md5.sh 指定要发送文件的路径`)
	SSH.Flag.StringVar(&lgInfoCFG.dpath, "d", "", `-d="/tmp/md5.sh" 指定文件保存路径`)
	SSH.Flag.StringVar(&lgInfoCFG.user, "u", "", `-u="root" 指定登录的用户`)
	SSH.Flag.StringVar(&lgInfoCFG.passwd, "p", "", `-p="passwd" 指定登录用户密码`)
	SSH.Flag.StringVar(&lgInfoCFG.out, "o", "", `-o="result" 指定结果输出文件`)
}

func lginfo(cmd *Command, arg []string) bool {
	var host []string
	var hosts [][]string

	if lgInfoCFG.hosts == "" && lgInfoCFG.config == "" {
		return false
	} else {
		if lgInfoCFG.hosts != "" {
			host = strings.Split(lgInfoCFG.hosts, ",")
		} else {
			var err error
			hosts, err = FileLine(lgInfoCFG.config, 3)
			if err != nil {
				log.Println("读取主机列表失败,", err.Error())
				return true
			}
		}
		if len(host) <= 0 && len(hosts) <= 0 {
			return false
		}
	}

	var output = os.Stdout
	defer output.Close()

	if lgInfoCFG.out != "" {
		var err error
		output, err = os.Create(lgInfoCFG.out)
		if err != nil {
			log.Println("创建结果文件失败:", err)
			return true
		}
	}
	if lgInfoCFG.cmd == "" && (lgInfoCFG.sfile == "" || lgInfoCFG.dpath == "") {
		return false
	}

	wait := new(sync.WaitGroup)
	if host != nil {
		if lgInfoCFG.user == "" || lgInfoCFG.passwd == "" {
			return false
		}

		for _, h := range host {
			c := newsshInfo(lgInfoCFG.user, lgInfoCFG.passwd, h)
			wait.Add(1)
			if lgInfoCFG.cmd != "" {
				go sendcommand(lgInfoCFG.cmd, wait, c, output)
			} else {
				go sendfile(lgInfoCFG.sfile, lgInfoCFG.dpath, wait, c, output)
			}
		}
	} else {
		for _, info := range hosts {
			c := newsshInfo(info[1], info[2], info[0])
			wait.Add(1)
			if lgInfoCFG.cmd != "" {
				go sendcommand(lgInfoCFG.cmd, wait, c, output)
			} else {
				go sendfile(lgInfoCFG.sfile, lgInfoCFG.dpath, wait, c, output)
			}
		}
	}
	wait.Wait()
	return true
}

func sendcommand(cmd string, wait *sync.WaitGroup, c *sshInfof, w io.Writer) {
	defer wait.Done()
	buf, err := c.SendCommand(cmd)
	if err != nil {
		if _, ok := err.(*net.OpError); ok {
			fmt.Fprintf(w, "连接:%s失败,错误信息:%s\n", c.host, err.Error())
		} else {
			fmt.Fprintf(w, "主机:%s执行命令失败,错误信息:%s\n", c.host, err.Error())
		}
		return
	}

	fmt.Fprintf(w, "%s执行结果:\n%s\n", c.host, string(buf))
}
func sendfile(sfile, dpath string, wait *sync.WaitGroup, c *sshInfof, w io.Writer) {
	defer wait.Done()
	err := c.SendFile(sfile, dpath)
	if err != nil {
		if _, ok := err.(*net.OpError); ok {
			fmt.Fprintf(w, "连接:%s失败,错误信息:%s\n", c.host, err.Error())
		} else {
			fmt.Fprintf(w, "主机:%s发送文件失败,错误信息:%s\n", c.host, err.Error())
		}
	}
	fmt.Fprintf(w, "%s发送文件成功 \n", c.host)
}

func newsshInfo(user, passwd, host string) *sshInfof {
	cfg := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(passwd),
		},
	}
	return &sshInfof{host: host, config: cfg}
}

type sshInfof struct {
	host   string
	client *ssh.Client
	config *ssh.ClientConfig
}

func (info *sshInfof) Dial() (err error) {
	info.client, err = ssh.DialTimeOut("tcp", info.host, 30, info.config)
	return
}

func (info *sshInfof) SendCommand(cmd string) ([]byte, error) {
	if info.client == nil {
		if err := info.Dial(); err != nil {
			return nil, err
		}
	}
	session, err := info.client.NewSession()
	if err != nil {
		return nil, err
	}
	return session.CombinedOutput(cmd)
}

func (info *sshInfof) SendFile(file, dirpath string) error {
	if info.client == nil {
		if err := info.Dial(); err != nil {
			return err
		}
	}

	File, err := os.Open(file)
	if err != nil {
		return err
	}

	defer File.Close()
	session, err := info.client.NewSession()
	if err != nil {
		return err
	}

	defer session.Close()
	stat, _ := File.Stat()

	go func() {
		w, _ := session.StdinPipe()
		fmt.Fprintln(w, "C0644", stat.Size(), filepath.Base(File.Name()))
		io.CopyN(w, File, stat.Size())
		fmt.Fprint(w, "\x00")
		w.Close()
	}()

	err = session.Run(fmt.Sprintf("/usr/bin/scp -qrt %s", dirpath))
	return err
}

func (info *sshInfof) Close() {
	if info.client != nil {
		info.client.Close()
	}
}
