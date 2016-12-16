package command

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var Commands = []*Command{
	FileServer, Mail,
	Compress, Network,
	Find, Md5,
	Compare, Fsnotify,
	SSH, FTP, Replace,
}

type Command struct {
	Run       func(cmd *Command, args []string) bool
	UsageLine string
	Short     string
	Long      string
	Flag      flag.FlagSet
	IsDebug   *bool
}

func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func (c *Command) Usage() {
	fmt.Fprintf(os.Stderr, "Example: %s %s\n", os.Args[0], c.UsageLine)
	fmt.Fprintf(os.Stderr, "Default Usage:\n")
	c.Flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "Description:\n")
	fmt.Fprintf(os.Stderr, "\t%s\n", strings.TrimSpace(c.Long))
	os.Exit(1)
}

func (c *Command) Runnable() bool {
	return c.Run != nil
}

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
		l := split(string(line))
		if len(l) == count {
			list = append(list, l)
		} else {
			log.Printf("无效的数据:%s\n", string(line))
		}
	}
}

func split(str string) []string {
	var l []string
	list := strings.Split(str, " ")
	for _, v := range list {
		if len(v) == 0 {
			continue
		}
		if strings.Contains(v, "	") {
			list := strings.Split(v, "	")
			for _, v := range list {
				if len(v) == 0 {
					continue
				}
				l = append(l, v)
			}
			continue
		}
		l = append(l, v)
	}
	return l
}
