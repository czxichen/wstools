package command

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
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

var Net = &cobra.Command{
	Use: "net",
	Example: `	并发模式ping测试指定主机的网络连通性
	-a ping -i www.baidu.com,www.163.com -c 4 -q
	telnet测试指定端口是否开启
	-a telnet -i www.baidu.com:80`,
	Short: "网络连通性工具",
	Long:  "网络连通性工具,支持ping|telnet命令",
	RunE:  net_run,
}

type net_config struct {
	Action  string
	Host    string
	Hosts   string
	Sum     bool
	Quick   bool
	Count   int
	TimeOut int
}

var _net net_config

func init() {
	Net.PersistentFlags().StringVarP(&_net.Action, "action", "a", "ping", "指定动作只支持ping,telnet两种")
	Net.PersistentFlags().StringVarP(&_net.Host, "ip", "i", "", "指定目标地址,当-a为telnet的时候远程地址必须包含端口,多地址用','分割")
	Net.PersistentFlags().StringVarP(&_net.Hosts, "hosts", "H", "", "文件读取目标地址,按行解析,如果指定-h则此参数无效")
	Net.PersistentFlags().IntVarP(&_net.TimeOut, "timeout", "t", 5, "设置超时时间,使用telnet的时候此参数有效")
	Net.PersistentFlags().IntVarP(&_net.Count, "count", "c", 2, "指定发出ping的次数")
	Net.PersistentFlags().BoolVarP(&_net.Sum, "sum", "s", false, "以统计方式输出结果")
	Net.PersistentFlags().BoolVarP(&_net.Quick, "quick", "q", false, "使用并发模式")
}

func net_run(cmd *cobra.Command, args []string) error {
	if _net.Hosts == "" && _net.Host == "" {
		return fmt.Errorf("必须指定-h或-H参数")
	}
	var list []string
	if _net.Host != "" {
		list = strings.Split(_net.Host, ",")
	} else {
		File, err := os.Open(_net.Hosts)
		if err != nil {
			fmt.Printf("读取主机列表出错:%s\n", err.Error())
			return nil
		}
		buf := bufio.NewReader(File)
		for {
			line, _, err := buf.ReadLine()
			if err != nil {
				break
			}
			if host := strings.TrimSpace(string(line)); host != "" {
				list = append(list, host)
			}
		}
		if len(list) == 0 {
			return nil
		}
	}
	var data = []byte("abcdefghijklmnopqrstuvwabcdefghi")
	var wait = new(sync.WaitGroup)

	switch _net.Action {
	case "ping":
		for _, host := range list {
			p, err := newPing(host, 8, data)
			if err != nil {
				fmt.Printf("Ping %s faild,%s\n", host, err.Error())
				continue
			}
			var Ping func(c int)
			if _net.Sum {
				Ping = p.PingCount
			} else {
				Ping = p.Ping
			}
			if _net.Quick {
				wait.Add(1)
				go func() {
					Ping(_net.Count)
					wait.Done()
				}()
			} else {
				Ping(_net.Count)
			}
		}
		wait.Wait()
	case "telnet":
		for _, host := range list {
			if _net.Quick {
				wait.Add(1)
				go func(host string) {
					if portIsOpen(host, _net.TimeOut) {
						fmt.Printf("Host:%s telnet sucess\n", host)
					} else {
						fmt.Printf("Host:%s telnet faild\n", host)
					}
					wait.Done()
				}(host)
			} else {
				if portIsOpen(host, _net.TimeOut) {
					fmt.Printf("Host:%s telnet sucess\n", host)
				} else {
					fmt.Printf("Host:%s telnet faild\n", host)
				}
			}
		}
		wait.Wait()
	default:
		return nil
	}
	return nil
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
	var buf = bytes.NewBuffer(nil)
	if err := self.Dail(); err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("Start ping from %s\n", self.Addr)
	for i := 0; i < count; i++ {
		self.SetDeadline(3)
		r := sendPingMsg(self.Conn, self.Data)
		if r.Error != nil {
			if opt, ok := r.Error.(*net.OpError); ok && opt.Timeout() {
				fmt.Fprintf(buf, "From %s reply: TimeOut\n", self.Addr)
				if err := self.Dail(); err != nil {
					fmt.Fprintf(buf, "Not found remote host\n")
					break
				}
			} else {
				fmt.Fprintf(buf, "From %s reply: %s\n", self.Addr, r.Error)
			}
		} else {
			fmt.Fprintf(buf, "From %s reply: bytes=32 time=%dms ttl=%d\n", self.Addr, r.Time, r.TTL)
		}
		time.Sleep(1e9)
	}
	if buf != nil {
		fmt.Println(string(buf.Bytes()))
	}
}

func (self *ping) PingCount(count int) {
	if err := self.Dail(); err != nil {
		fmt.Println(err.Error())
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
	fmt.Printf("From %s reply:sucess=%d abytes=32 atime=%.2fms attl=%.2f faild=%d\n",
		self.Addr, sucess, float64(times)/float64(sucess), float64(ttl)/float64(sucess), errs)
}

func (self *ping) Dail() (err error) {
	self.Conn, err = net.Dial("ip4:icmp", self.Addr)
	return
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
