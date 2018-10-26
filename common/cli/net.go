package cli

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

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// NetConfig net 命令配置
type NetConfig struct {
	Action  string
	Host    string
	Hosts   string
	Sum     bool
	Quick   bool
	Count   int
	TimeOut int
}

// NetRun 运行网络工具
func NetRun(netConfig *NetConfig) error {
	if netConfig.Hosts == "" && netConfig.Host == "" {
		return fmt.Errorf("必须指定-h或-H参数")
	}
	var list []string
	if netConfig.Host != "" {
		list = strings.Split(netConfig.Host, ",")
	} else {
		File, err := os.Open(netConfig.Hosts)
		if err != nil {
			return err
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

	switch netConfig.Action {
	case "ping":
		for _, host := range list {
			p, err := NetPingNew(host, 8, data)
			if err != nil {
				fmt.Printf("Ping %s faild,%s\n", host, err.Error())
				continue
			}
			var Ping func(c int)
			if netConfig.Sum {
				Ping = p.PingCount
			} else {
				Ping = p.Ping
			}
			if netConfig.Quick {
				wait.Add(1)
				go func() {
					Ping(netConfig.Count)
					wait.Done()
				}()
			} else {
				Ping(netConfig.Count)
			}
		}
		wait.Wait()
	case "telnet":
		for _, host := range list {
			if netConfig.Quick {
				wait.Add(1)
				go func(host string) {
					if portIsOpen(host, netConfig.TimeOut) {
						fmt.Printf("Host:%s telnet sucess\n", host)
					} else {
						fmt.Printf("Host:%s telnet faild\n", host)
					}
					wait.Done()
				}(host)
			} else {
				if portIsOpen(host, netConfig.TimeOut) {
					fmt.Printf("Host:%s telnet sucess\n", host)
				} else {
					fmt.Printf("Host:%s telnet faild\n", host)
				}
			}
		}
		wait.Wait()
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

// NetPingNew 创建net ping
func NetPingNew(addr string, req int, data []byte) (*NetPing, error) {
	wb, err := marshalMsg(req, data)
	if err != nil {
		return nil, err
	}
	addr, err = lookup(addr)
	if err != nil {
		return nil, err
	}
	return &NetPing{Data: wb, Addr: addr}, nil
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
	// 从解析的地址随机一个IP进行访问
	return addrs[rd.Intn(len(addrs))], nil
}

// marshalMsg 封装消息
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

// NetPing net ping
type NetPing struct {
	conn net.Conn
	Addr string
	Data []byte
}

// Ping ping
func (p *NetPing) Ping(count int) {
	var buf = bytes.NewBuffer(nil)
	if err := p.Init(); err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("Start ping from %s\n", p.Addr)
	for i := 0; i < count; i++ {
		p.SetDeadline(3)
		r := sendPingMsg(p.conn, p.Data)
		if r.Error != nil {
			if opt, ok := r.Error.(*net.OpError); ok && opt.Timeout() {
				fmt.Fprintf(buf, "From %s Reply: TimeOut\n", p.Addr)
				if err := p.Init(); err != nil {
					fmt.Fprintf(buf, "Not found remote host\n")
					break
				}
			} else {
				fmt.Fprintf(buf, "From %s Reply: %s\n", p.Addr, r.Error)
			}
		} else {
			fmt.Fprintf(buf, "From %s Reply: bytes=32 time=%dms ttl=%d\n", p.Addr, r.Time, r.TTL)
		}
		time.Sleep(1e9)
	}
	if buf != nil {
		fmt.Println(string(buf.Bytes()))
	}
}

// PingCount ping统计
func (p *NetPing) PingCount(count int) {
	if err := p.Init(); err != nil {
		fmt.Println(err.Error())
		return
	}

	var times, ttl, errs int
	for i := 0; i < count; i++ {
		p.SetDeadline(3)
		r := sendPingMsg(p.conn, p.Data)
		if r.Error != nil {
			errs++
			continue
		}
		times += int(r.Time)
		ttl += int(r.TTL)
		time.Sleep(1e9)
	}
	sucess := count - errs
	fmt.Printf("From %s Reply:sucess=%d abytes=32 atime=%.2fms attl=%.2f faild=%d\n",
		p.Addr, sucess, float64(times)/float64(sucess), float64(ttl)/float64(sucess), errs)
}

// Init 初始化Ping
func (p *NetPing) Init() (err error) {
	p.conn, err = net.Dial("ip4:icmp", p.Addr)
	return
}

// SetDeadline 设置超时时间单位s
func (p *NetPing) SetDeadline(timeout int) error {
	return p.conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
}

// Close 关闭远程连接
func (p *NetPing) Close() error {
	return p.conn.Close()
}

// reply 响应信息
type reply struct {
	Time  int64
	TTL   uint8
	Error error
}

func sendPingMsg(c net.Conn, data []byte) (rep *reply) {
	start := time.Now()
	if _, rep.Error = c.Write(data); rep.Error != nil {
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
		rep = &reply{t, ttl, nil}
	case ipv4.ICMPTypeDestinationUnreachable:
		rep.Error = errors.New("Destination Unreachable")
	default:
		rep.Error = fmt.Errorf("Not ICMPTypeEchoReply %v", rm)
	}
	return
}
