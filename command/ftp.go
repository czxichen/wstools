package command

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var Ftp = &cobra.Command{
	Use: "ftp",
	Example: `	下载远程/tmp/main.go文件到./目录下
	-D -u root -p toor -d /server/Res/scenes.ini -s nima -H  127.0.0.1:21
	上传文件到/Server/main.go
	-u root -p toor -s main.go -d /Server/main.go -H 127.0.0.1:21`,
	Short: "FTP上传下载",
	Long:  "使用简单的FTP协议实现文件的上传下载",
	RunE:  ftp_run,
}

type ftp_config struct {
	Host        string
	User        string
	Passwd      string
	Source      string
	Destination string
	Download    bool
}

var _ftp ftp_config

func init() {
	Ftp.PersistentFlags().StringVarP(&_ftp.Host, "host", "H", "", "指定ftp地址端口,不能为空")
	Ftp.PersistentFlags().StringVarP(&_ftp.User, "user", "u", "", "指定登录的用户名,不能为空")
	Ftp.PersistentFlags().StringVarP(&_ftp.Passwd, "passwd", "p", "", "指定登录的用户密码,不能为空")
	Ftp.PersistentFlags().StringVarP(&_ftp.Source, "source", "s", "", "指定原始文件路径,不能为空")
	Ftp.PersistentFlags().StringVarP(&_ftp.Destination, "destination", "d", "", "指定目标文件路径,不能为空")
	Ftp.PersistentFlags().BoolVarP(&_ftp.Download, "download", "D", false, "从ftp上下载文件")
}

func ftp_run(cmd *cobra.Command, args []string) error {
	if _ftp.Host == "" || _ftp.User == "" || _ftp.Passwd == "" || _ftp.Source == "" || _ftp.Destination == "" {
		return fmt.Errorf("参数错误")
	}

	conn, err := newFTP(_ftp.Host, _ftp.User, _ftp.Passwd)
	if err != nil {
		fmt.Printf("登录失败:%s\n", err.Error())
		return nil
	}
	if _ftp.Download {
		err = conn.GetFile(_ftp.Source, _ftp.Destination)
	} else {
		err = conn.PutFile(_ftp.Source, _ftp.Destination)
	}
	if err != nil {
		fmt.Printf("文件传输失败:%s\n", err.Error())
	}
	return nil
}

func newFTP(ip, user, pass string) (*ftplogin, error) {
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
			fmt.Print(string(line))
		}
		if bytes.HasPrefix(line, []byte("230")) {
			fmt.Printf(string(line))
			break
		}
	}
	return &ftplogin{con, ip}, nil
}

type ftplogin struct {
	con net.Conn
	ip  string
}

func (self *ftplogin) PutFile(lpath, rpath string) error {
	con, err := self.connection("STOR", rpath)
	if err != nil {
		return err
	}
	defer con.Close()
	File, err := os.Open(lpath)
	if err != nil {
		return err
	}
	defer File.Close()
	_, err = io.Copy(con, File)
	if err != nil {
		return err
	}

	buf := make([]byte, 1024)
	n, err := self.con.Read(buf)
	if err == nil {
		fmt.Print(string(buf[:n]))
	}
	return err
}

func (self *ftplogin) GetFile(lpath, rpath string) error {
	con, err := self.connection("RETR", rpath)
	if err != nil {
		return err
	}

	defer con.Close()
	File, err := os.Create(lpath)
	if err != nil {
		return err
	}

	defer File.Close()
	_, err = io.Copy(File, con)
	if err != nil {
		return err
	}

	buf := make([]byte, 1024)
	n, err := self.con.Read(buf)
	if err == nil {
		fmt.Print(string(buf[:n]))
	}
	return err
}

func (self *ftplogin) connection(status, Pathname string) (net.Conn, error) {
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
	defer con.Close()
	fmt.Fprintf(self.con, "%s %s\r\n", status, Pathname)
	n, err = self.con.Read(buf)
	if err != nil {
		return nil, err
	}
	if !strings.Contains(string(buf[:n]), "150 Opening data channel") {
		return nil, errors.New(string(buf[:n-2]))
	}
	return con, nil
}

func (self *ftplogin) Exit() {
	buf := make([]byte, 1024)
	fmt.Fprintln(self.con, "QUIT ")
	n, err := self.con.Read(buf)
	if err == nil {
		fmt.Print(string(buf[:n]))
	} else {
		fmt.Println(err)
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
