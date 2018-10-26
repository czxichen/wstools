package cli

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHDial 创建链接
func SSHDial(address, user string, auth []ssh.AuthMethod, timeout int) (*ssh.Client, error) {
	cliConfig := &ssh.ClientConfig{
		User:            user,
		Auth:            auth,
		Timeout:         time.Second * time.Duration(timeout),
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
	}
	return ssh.Dial("tcp", address, cliConfig)
}

// SSHAuth 获取认证信息
func SSHAuth(passwd string, privateKey ...string) ([]ssh.AuthMethod, error) {
	var auths = make([]ssh.AuthMethod, 0, 2)
	if passwd != "" {
		auths = append(auths, ssh.Password(passwd))
	}

	var sigs = make([]ssh.Signer, len(privateKey))
	for idx, keyPath := range privateKey {
		key, err := ioutil.ReadFile(keyPath)
		if err != nil {
			return nil, err
		}
		sig, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}
		sigs[idx] = sig
	}
	return append(auths, ssh.PublicKeys(sigs...)), nil
}

// SSHSendFile 发送文件
func SSHSendFile(cli *ssh.Client, srcPath, dstPath string) error {
	File, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer File.Close()

	session, err := cli.NewSession()
	if err != nil {
		return nil
	}
	defer session.Close()

	stat, err := File.Stat()
	if err != nil {
		return err
	}
	go func() {
		dstFile, err := session.StdinPipe()
		if err == nil {
			fmt.Fprintln(dstFile, "C0644", stat.Size(), filepath.Base(dstPath))
			io.CopyN(dstFile, File, stat.Size())
			fmt.Fprint(dstFile, "\x00")
			dstFile.Close()
		}
	}()
	return session.Run(fmt.Sprintf("/usr/bin/scp -qrt %s", dstPath))
}

// SSHBatchSendFile 批量发送文件
func SSHBatchSendFile(conns []*SSHConnection, timeout int, srcPath, dstPath string, output io.Writer) {
	clients := InitClients(conns, timeout, output)
	if len(clients) == 0 {
		return
	}

	var resultChan = make(chan *SSHResult, 1)
	for host, cli := range clients {
		go func(host string, client *CMDClient) {
			var result = &SSHResult{Host: host}
			result.Error = SSHSendFile(client.Client, srcPath, dstPath)
			resultChan <- result
		}(host, cli)
	}

	for i := 0; i < len(clients); i++ {
		result := <-resultChan
		if result.Error == nil {
			fmt.Fprintf(output, "[INFO] 发送成功:%s\n", result.Host)
		} else {
			fmt.Fprintf(output, "[ERROR] 发送失败:%v\n", result.Error)
		}
	}
	close(resultChan)
}

// SSHSendCommondOnce 发送单次单个命令
func SSHSendCommondOnce(cli *ssh.Client, cmd string, output io.Writer) error {
	session, err := cli.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	session.Stdout = output
	session.Stderr = output
	return session.Run(cmd)
}

// SSHSendCommonds 发送命令
func SSHSendCommonds(host string, cli *ssh.Client, cmdChan <-chan string, outChan chan<- *SSHResult) error {
	session, err := cli.NewSession()
	if err != nil {
		return err
	}
	go func() {
		var ret = &SSHResult{Host: host}
		for cmd := range cmdChan {
			result, err := session.Output(cmd)
			if err == nil {
				ret.Data = result
			} else {
				ret.Error = err
			}
			outChan <- ret
		}
		session.Close()
	}()
	return nil
}

// SSHBatchCommond 批量发送命令
func SSHBatchCommond(ctx context.Context, conns []*SSHConnection, timeout int, cmdChan <-chan string, output io.Writer) {
	var resultChan = make(chan *SSHResult, len(conns))
	var clients = InitClients(conns, timeout, output)
	for host, client := range clients {
		if err := SSHSendCommonds(host, client.Client, client.CMDChan, resultChan); err != nil {
			fmt.Fprintf(output, "[ERROR] 创建会话失败:%s\n", err.Error())
		}
	}
	if len(clients) == 0 {
		return
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case result := <-resultChan:
				if result.Error == nil {
					fmt.Fprintf(output, "---------------------------SUCCESS\t%s---------------------------\n%s\n------------------------------------------------------\n", result.Host, result.Data)
				} else {
					fmt.Fprintf(output, "---------------------------FAILD\t%s---------------------------\n%s\n------------------------------------------------------\n", result.Host, result.Error.Error())
				}
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			for _, client := range clients {
				client.Close()
			}
			close(resultChan)
			return
		case cmd := <-cmdChan:
			for _, cli := range clients {
				cli.CMDChan <- cmd
			}
		}
	}

}

// SSHShellCommond 发送命令
func SSHShellCommond(cli *ssh.Client, read io.Reader, output io.Writer) error {
	session, err := cli.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdin = read
	session.Stdout = output
	session.Stderr = output

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err = session.RequestPty("xterm", 80, 320, modes); err != nil {
		return err
	}
	if err = session.Shell(); err != nil {
		return err
	}
	return session.Wait()
}

// InitClients 初始化客户端
func InitClients(conns []*SSHConnection, timeout int, output io.Writer) map[string]*CMDClient {
	var clients = make(map[string]*CMDClient, len(conns))
	for _, conn := range conns {
		auth, err := SSHAuth(conn.Passwd, conn.Keys...)
		if err != nil {
			fmt.Fprintf(output, "[ERROR] 认证解析错误:%s\n", err.Error())
			continue
		}
		client, err := SSHDial(conn.Host, conn.User, auth, timeout)
		if err != nil {
			fmt.Fprintf(output, "[ERROR] 创建连接失败:%s\n", err.Error())
			continue
		}
		clients[conn.Host] = &CMDClient{Client: client, CMDChan: make(chan string)}
	}
	return clients
}

// SSHConnection 连接信息
type SSHConnection struct {
	Host   string   `json:"host"`
	User   string   `json:"user"`
	Passwd string   `json:"passwd"`
	Keys   []string `json:"keys"`
}

// SSHResult 返回数据
type SSHResult struct {
	Host  string `json:"host"`
	Data  []byte `json:"data"`
	Error error  `json:"error"`
}

// CMDClient 客户端
type CMDClient struct {
	Client  *ssh.Client
	CMDChan chan string
}

// Close 关闭client
func (cc *CMDClient) Close() error {
	close(cc.CMDChan)
	return cc.Client.Close()
}
