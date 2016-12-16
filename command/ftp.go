package command

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type ftpcfg struct {
	get    bool
	host   string
	user   string
	rpath  string
	lpath  string
	passwd string
}

var (
	ftpCFG = ftpcfg{}
	FTP    = &Command{
		UsageLine: `ftp -l main.go -r /mnt/main.go`,
		Run:       ftpclient,
		Short:     "使用ftp协议下载或上传文件",
		Long: `实现用的是PASV模式,上传或这下载文件路径不能是目录,
	wstools ftp -l main.go -r /mnt/main.go
	wstools ftp -l main.go -r /mnt/main.go -g false
`,
	}
)

func init() {
	FTP.Flag.StringVar(&ftpCFG.host, "H", "", `-H="127.0.0.1:21" 指定ftp主机地址`)
	FTP.Flag.StringVar(&ftpCFG.user, "u", "", `-u="root" 指定登录的用户名称`)
	FTP.Flag.StringVar(&ftpCFG.passwd, "p", "", `-p=" 指定登录用户的密码"`)
	FTP.Flag.StringVar(&ftpCFG.lpath, "l", "", `-l="main.go" 本地文件路径`)
	FTP.Flag.StringVar(&ftpCFG.rpath, "r", "", `-d="/main.go" 远程文件路径`)
	FTP.Flag.BoolVar(&ftpCFG.get, "g", true, `-g=false 当指定值为false的时候,表示上传文件`)
}

func ftpclient(cmd *Command, args []string) bool {
	if ftpCFG.lpath == "" || ftpCFG.rpath == "" {
		return false
	}

	if ftpCFG.host != "" && ftpCFG.user != "" && ftpCFG.passwd != "" {
		f, err := newFTP(ftpCFG.host, ftpCFG.user, ftpCFG.passwd)
		if err == nil {
			defer f.Exit()
			if ftpCFG.get {
				err = f.GetFile(ftpCFG.lpath, ftpCFG.rpath)
			} else {
				err = f.PutFile(ftpCFG.lpath, ftpCFG.rpath)
			}
		}
		if err != nil {
			log.Println(err)
		}
		return true
	}

	return false
}

func newFTP(ip, user, pass string) (*ftp, error) {
	con, err := net.Dial("tcp", ip)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(con, "USER %s\r\nPASS %s\r\n", user, pass)
	buf := bufio.NewReader(con)
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			return nil, err
		}
		if bytes.HasPrefix(line, []byte("220")) {
			continue
		}
		if bytes.HasPrefix(line, []byte("530")) {
			log.Print(string(line))
		}
		if bytes.HasPrefix(line, []byte("230")) {
			log.Printf(string(line))
			break
		}
	}
	return &ftp{con, ip}, nil
}

type ftp struct {
	con net.Conn
	ip  string
}

func (self *ftp) PutFile(lpath, rpath string) error {
	con, err := self.connection("STOR", rpath)
	if err != nil {
		return err
	}
	File, err := os.Open(lpath)
	if err != nil {
		con.Close()
		return err
	}
	io.Copy(con, File)
	con.Close()
	File.Close()

	buf := make([]byte, 1024)
	n, err := self.con.Read(buf)
	if err == nil {
		log.Print(string(buf[:n]))
	}
	return err
}

func (self *ftp) GetFile(lpath, rpath string) error {
	con, err := self.connection("RETR", rpath)
	if err != nil {
		return err
	}
	File, err := os.Create(lpath)
	if err != nil {
		con.Close()
		return err
	}
	io.Copy(File, con)
	File.Close()
	con.Close()
	buf := make([]byte, 1024)
	n, err := self.con.Read(buf)
	if err == nil {
		log.Print(string(buf[:n]))
	}
	return err
}

func (self *ftp) connection(status, Pathname string) (net.Conn, error) {
	buf := make([]byte, 1024)
	fmt.Fprintln(self.con, "PASV ")
	n, err := self.con.Read(buf)
	if err != nil {
		return nil, err
	}
	if s := string(buf[:n]); !strings.Contains(s, "227 Entering Passive Mode") {
		return nil, errors.New(s)
	}
	port := getport(buf[27 : n-3])
	con, err := net.Dial("tcp", fmt.Sprintf("%s:%d", strings.Split(self.ip, ":")[0], port))
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(self.con, "%s %s\r\n", status, Pathname)
	n, err = self.con.Read(buf)
	if err != nil {
		con.Close()
		return nil, err
	}
	if !strings.Contains(string(buf[:n]), "150 Opening data channel") {
		con.Close()
		return nil, errors.New(string(buf[:n-2]))
	}
	return con, nil
}

func (self *ftp) Exit() {
	buf := make([]byte, 1024)
	fmt.Fprintln(self.con, "QUIT ")
	n, err := self.con.Read(buf)
	if err == nil {
		log.Print(string(buf[:n]))
	} else {
		log.Println(err)
	}
	if self.con != nil {
		self.con.Close()
	}
}

func getport(by []byte) int {
	s := string(by)
	list := strings.Split(s, ",")
	n1, err := strconv.Atoi(list[len(list)-2])
	if err != nil {
		return 0
	}
	n2, err := strconv.Atoi(list[len(list)-1])
	if err != nil {
		return 0
	}
	return n1*256 + n2
}
