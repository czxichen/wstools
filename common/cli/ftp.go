package cli

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
)

// FTPClient ftp client interface
type FTPClient interface {
	PutFile(lpath, rpath string) error
	GetFile(lpath, rpath string) error
	Exit()
}

// NewFTP 创建ftp连接
func NewFTP(ip, user, pass string) (FTPClient, error) {
	conn, err := net.Dial("tcp", ip)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(conn, "USER %s\r\nPASS %s\r\n", user, pass)
	buf := bufio.NewReader(conn)
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
	return &ftpLogin{ip, conn}, nil
}

// ftpLogin ftpLogin
type ftpLogin struct {
	ip   string
	conn net.Conn
}

// PutFile 上传文件
func (login *ftpLogin) PutFile(lpath, rpath string) error {
	c, err := login.connection("STOR", rpath)
	if err != nil {
		return err
	}
	defer c.Close()
	File, err := os.Open(lpath)
	if err != nil {
		return err
	}
	defer File.Close()
	_, err = io.Copy(c, File)
	if err != nil {
		return err
	}

	buf := make([]byte, 1024)
	n, err := c.Read(buf)
	if err == nil {
		fmt.Print(string(buf[:n]))
	}
	return err
}

// GetFile 下载文件
func (login *ftpLogin) GetFile(lpath, rpath string) error {
	con, err := login.connection("RETR", rpath)
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
	n, err := login.conn.Read(buf)
	if err == nil {
		fmt.Print(string(buf[:n]))
	}
	return err
}

func (login *ftpLogin) connection(status, Pathname string) (net.Conn, error) {
	buf := make([]byte, 1024)
	fmt.Fprintln(login.conn, "PASV ")
	n, err := login.conn.Read(buf)
	if err != nil {
		return nil, err
	}
	if s := string(buf[:n]); !strings.Contains(s, "227 Entering Passive Mode") {
		return nil, errors.New(s)
	}
	port := getport(buf[27 : n-3])
	con, err := net.Dial("tcp", fmt.Sprintf("%s:%d", strings.Split(login.ip, ":")[0], port))
	if err != nil {
		return nil, err
	}
	defer con.Close()
	fmt.Fprintf(login.conn, "%s %s\r\n", status, Pathname)
	n, err = login.conn.Read(buf)
	if err != nil {
		return nil, err
	}
	if !strings.Contains(string(buf[:n]), "150 Opening data channel") {
		return nil, errors.New(string(buf[:n-2]))
	}
	return con, nil
}

// Exit 退出
func (login *ftpLogin) Exit() {
	buf := make([]byte, 1024)
	fmt.Fprintln(login.conn, "QUIT ")
	n, err := login.conn.Read(buf)
	if err == nil {
		fmt.Print(string(buf[:n]))
	} else {
		fmt.Println(err)
	}
	if login.conn != nil {
		login.conn.Close()
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
