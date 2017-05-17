package command

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type netconfig struct {
	Action    string
	Host      string
	TimeOut   int
	Count     bool
	PingCount int
	File      string
	QuickMode bool
}

var (
	networkCFG netconfig
	Network    = &Command{
		UsageLine: `net -a telnet -i 127.0.0.1 -t 10`,
		Run:       network,
		Short:     "检测远程地址或端口是否通",
		Long: `可以单独或者个批量操作多个远程地址,测试远程地址的连通性
	wstools net -H 127.0.0.1,www.baidu.com
	wstools net -H 127.0.0.1,www.baidu.com -C 4 -c -q
	wstools net -a telnet -H 127.0.0.1:80,www.baidu.com:80
`,
	}
)

func init() {
	Network.Flag.StringVar(&networkCFG.Action, "a", "ping", `-a="ping" 指定测试方式,只支持ping,telnet两种`)
	Network.Flag.StringVar(&networkCFG.Host, "H", "", `-H="127.0.0.1" 指定目标地址,当-a为telnet的时候远程地址必须包含端口,多地址用','分割`)
	Network.Flag.IntVar(&networkCFG.TimeOut, "t", 5, `-t=5 设置超时时间,使用telnet的时候此参数有效`)
	Network.Flag.BoolVar(&networkCFG.Count, "c", false, `-c=true 以统计方式输出结果`)
	Network.Flag.BoolVar(&networkCFG.QuickMode, "q", false, `-q=true 使用并发模式`)
	Network.Flag.IntVar(&networkCFG.PingCount, "C", 2, `-C=2 指定发出ping的次数`)
	Network.Flag.StringVar(&networkCFG.File, "F", "", `-F=iplist.txt 从文件读取目标地址,每行一个,如果指定-H则此参数无效`)
}

func network(cmd *Command, args []string) bool {
	if networkCFG.File == "" && networkCFG.Host == "" {
		return false
	}
	var list []string
	if networkCFG.Host != "" {
		list = strings.Split(networkCFG.Host, ",")
	} else {
		File, err := os.Open(networkCFG.File)
		if err != nil {
			log.Println(err)
			return true
		}
		buf := bufio.NewReader(File)
		for {
			line, _, err := buf.ReadLine()
			if err != nil {
				break
			}
			list = append(list, string(line))
		}
		if len(list) == 0 {
			log.Println("Host addr is null")
			return true
		}
	}
	var data = []byte("abcdefghijklmnopqrstuvwabcdefghi")
	var wait = new(sync.WaitGroup)
	switch networkCFG.Action {
	case "ping":
		for _, host := range list {
			p, err := newPing(host, 8, data)
			if err != nil {
				log.Printf("Ping %s faild,%s\n", host, err.Error())
				continue
			}
			var Ping func(c int)
			if networkCFG.Count {
				Ping = p.PingCount
			} else {
				Ping = p.Ping
			}
			if networkCFG.QuickMode {
				wait.Add(1)
				go func() {
					Ping(networkCFG.PingCount)
					wait.Done()
				}()
			} else {
				Ping(networkCFG.PingCount)
			}
		}
		wait.Wait()
	case "telnet":
		for _, host := range list {
			if networkCFG.QuickMode {
				wait.Add(1)
				go func(host string) {
					if portIsOpen(host, networkCFG.TimeOut) {
						log.Printf("Host:%s telnet sucess\n", host)
					} else {
						log.Printf("Host:%s telnet faild\n", host)
					}
					wait.Done()
				}(host)
			} else {
				if portIsOpen(host, networkCFG.TimeOut) {
					log.Printf("Host:%s telnet sucess\n", host)
				} else {
					log.Printf("Host:%s telnet faild\n", host)
				}
			}
		}
		wait.Wait()
	default:
		return false
	}
	return true
}

func portIsOpen(ip string, timeout int) bool {
	con, err := net.DialTimeout("tcp", ip, time.Duration(timeout)*time.Second)
	if err != nil {
		return false
	}
	con.Close()
	return true
}

func newPing(addr string, req int, data []byte) (*ping, error) {
	wb, err := marshalMsg(req, data)
	if err != nil {
		return nil, err
	}
	addr, err = lookup(addr)
	if err != nil {
		return nil, err
	}
	return &ping{Data: wb, Addr: addr}, nil
}

func lookup(host string) (string, error) {
	addrs, err := net.LookupHost(host)
	if err != nil {
		return "", err
	}
	if len(addrs) < 1 {
		return "", errors.New("unknown host")
	}
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	return addrs[rd.Intn(len(addrs))], nil
}

func marshalMsg(req int, data []byte) ([]byte, error) {
	xid, xseq := os.Getpid()&0xffff, req
	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: xid, Seq: xseq,
			Data: data,
		},
	}
	return wm.Marshal(nil)
}

type ping struct {
	Addr string
	Conn net.Conn
	Data []byte
}

func (self *ping) Ping(count int) {
	var buf *bytes.Buffer
	var logger = log.New(os.Stdout, "", log.LstdFlags)
	if networkCFG.QuickMode {
		buf = bytes.NewBuffer(nil)
		logger.SetOutput(buf)
	}

	if err := self.Dail(); err != nil {
		logger.Println("Not found remote host")
		return
	}
	log.Printf("Start ping from %s\n", self.Addr)
	for i := 0; i < count; i++ {
		self.SetDeadline(3)
		r := sendPingMsg(self.Conn, self.Data)
		if r.Error != nil {
			if opt, ok := r.Error.(*net.OpError); ok && opt.Timeout() {
				logger.Printf("From %s reply: TimeOut\n", self.Addr)
				if err := self.Dail(); err != nil {
					logger.Printf("Not found remote host\n")
					break
				}
			} else {
				logger.Printf("From %s reply: %s\n", self.Addr, r.Error)
			}
		} else {
			logger.Printf("From %s reply: bytes=32 time=%dms ttl=%d\n", self.Addr, r.Time, r.TTL)
		}
		time.Sleep(1e9)
	}
	if buf != nil {
		fmt.Println(string(buf.Bytes()))
	}
}

func (self *ping) PingCount(count int) {
	if err := self.Dail(); err != nil {
		log.Println(err.Error())
		return
	}

	var times, ttl, errs int
	for i := 0; i < count; i++ {
		self.SetDeadline(3)
		r := sendPingMsg(self.Conn, self.Data)
		if r.Error != nil {
			errs += 1
			continue
		}
		times += int(r.Time)
		ttl += int(r.TTL)
		time.Sleep(1e9)
	}
	sucess := count - errs
	log.Printf("From %s reply:sucess=%d abytes=32 atime=%.2fms attl=%.2f faild=%d",
		self.Addr, sucess, float64(times)/float64(sucess), float64(ttl)/float64(sucess), errs)
}

func (self *ping) Dail() (err error) {
	self.Conn, err = net.Dial("ip4:icmp", self.Addr)
	if err != nil {
		return err
	}
	return nil
}

//设置超时时间单位s
func (self *ping) SetDeadline(timeout int) error {
	return self.Conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
}

//关闭远程连接
func (self *ping) Close() error {
	return self.Conn.Close()
}

type reply struct {
	Time  int64
	TTL   uint8
	Error error
}

func sendPingMsg(c net.Conn, wb []byte) (rep reply) {
	start := time.Now()
	if _, rep.Error = c.Write(wb); rep.Error != nil {
		return
	}

	rb := make([]byte, 1500)
	var n int
	n, rep.Error = c.Read(rb)
	if rep.Error != nil {
		return
	}

	duration := time.Now().Sub(start)
	ttl := uint8(rb[8])
	rb = func(b []byte) []byte {
		if len(b) < 20 {
			return b
		}
		hdrlen := int(b[0]&0x0f) << 2
		return b[hdrlen:]
	}(rb)
	var rm *icmp.Message
	rm, rep.Error = icmp.ParseMessage(1, rb[:n])
	if rep.Error != nil {
		return
	}

	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		t := int64(duration / time.Millisecond)
		rep = reply{t, ttl, nil}
	case ipv4.ICMPTypeDestinationUnreachable:
		rep.Error = errors.New("Destination Unreachable")
	default:
		rep.Error = fmt.Errorf("Not ICMPTypeEchoReply %v", rm)
	}
	return
}
