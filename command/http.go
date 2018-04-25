package command

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/czxichen/command"
	"github.com/lucas-clemente/quic-go/h2quic"
	"github.com/spf13/cobra"
)

var Http = &cobra.Command{
	Use: "http",
	Example: `	开启目录访问uuid,并设置baseauth用户名和密码
	-d uuid -u root -p toor -i
	下载文件,并保存
	-w -H http://127.0.0.1:1789/type.proto -s t.proto -u root -p toor`,
	Short: "使用简单的http协议通讯",
	Long:  "使用http协议进行内容传输,支持文件上传下载",
	RunE:  http_run,
}

type http_config struct {
	Host        string
	User        string
	Passwd      string
	Crt         string
	Key         string
	Dir         string
	Save        string
	Wget        bool
	Quic        bool
	OnlyQuic    bool
	Index       bool
	Verbose     bool
	ForceVerify bool //当前不可用
}

var _http http_config

func init() {
	Http.PersistentFlags().StringVarP(&_http.Host, "host", "H", ":1789", "指定监听的地址端口,或者要访问的url")
	Http.PersistentFlags().StringVarP(&_http.User, "user", "u", "", "指定BaseAuth的用户名,可以为空")
	Http.PersistentFlags().StringVarP(&_http.Passwd, "passwd", "p", "", "指定BaseAuth的密码,可以为空")
	Http.PersistentFlags().StringVarP(&_http.Crt, "crt", "c", "", "指定TLS的Crt文件,可以为空")
	Http.PersistentFlags().StringVarP(&_http.Key, "key", "k", "", "指定TLS的Key文件,可以为空")
	Http.PersistentFlags().StringVarP(&_http.Dir, "dir", "d", "", "指定共享目录,当server启动的时候不能为空")
	Http.PersistentFlags().StringVarP(&_http.Save, "save", "s", "", "使用下载的时候,文件保存路径,为空则保存在当前目录")
	Http.PersistentFlags().BoolVarP(&_http.Wget, "wget", "w", false, "从指定的host下载文件")
	Http.PersistentFlags().BoolVarP(&_http.Quic, "quic", "q", false, "使用quic协议,默认会监听tcp,udp上")
	Http.PersistentFlags().BoolVarP(&_http.OnlyQuic, "onlyquic", "o", false, "仅启动quic协议,只监听在udp")
	Http.PersistentFlags().BoolVarP(&_http.Index, "index", "i", false, "启用目录索引,允许目录浏览")
	Http.PersistentFlags().BoolVarP(&_http.Verbose, "verbose", "v", true, "关闭后台访问输出")
}

func http_run(cmd *cobra.Command, args []string) error {
	var (
		err    error
		tlscfg *tls.Config
	)
	if _http.Wget {
		if strings.HasPrefix(_http.Host, "https://") || _http.Quic {
			tlscfg, err = parse_tls(_http)
			if err != nil {
				fmt.Printf("解析TLS文件失败:%s\n", err.Error())
				return nil
			}
		} else {
			if _http.Quic {
				fmt.Printf("必须使用https通信")
				return nil
			}
		}
		err = command.Wget(_http.Quic, _http.Host, _http.Save, _http.User, _http.Passwd, tlscfg)
		if err != nil {
			fmt.Printf("请求失败:%s\n", err.Error())
		}
		return nil
	}

	_http.Dir = filepath.Clean(_http.Dir)

	if _http.Crt != "" {
		if _http.Quic {
			if _http.OnlyQuic {
				err = h2quic.ListenAndServeQUIC(_http.Host, _http.Crt, _http.Key, _http)
			} else {
				err = h2quic.ListenAndServe(_http.Host, _http.Crt, _http.Key, _http)
			}
		} else {
			err = http.ListenAndServeTLS(_http.Host, _http.Crt, _http.Key, _http)
		}
	} else {
		err = http.ListenAndServe(_http.Host, _http)
	}
	if err != nil {
		fmt.Printf("Listen server error:%s\n", err.Error())
	}
	return nil
}

func (dir http_config) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if dir.Verbose {
		fmt.Printf("Remoter:%s\tRequest:%s\n", r.RemoteAddr, r.RequestURI)
	}

	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if dir.User != "" {
		user, pass, ok := r.BasicAuth()
		if !ok || user != dir.User || pass != dir.Passwd {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	var path = dir.Dir + r.RequestURI
	info, err := os.Lstat(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if info.IsDir() && !dir.Index {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, path)
}

func parse_tls(info http_config) (*tls.Config, error) {
	crts, err := tls.LoadX509KeyPair(info.Crt, info.Key)
	if err != nil {
		if os.IsNotExist(err) && info.Wget {
			return &tls.Config{InsecureSkipVerify: true}, nil
		}
		return nil, err
	}
	var pool = x509.NewCertPool()
	buf, err := ioutil.ReadFile(info.Crt)
	if err != nil {
		return nil, err
	}
	p := &pem.Block{}
	p, _ = pem.Decode(buf)
	crt, err := x509.ParseCertificate(p.Bytes)
	if err != nil {
		return nil, err
	}
	pool.AddCert(crt)

	var tlscfg = &tls.Config{
		Certificates: []tls.Certificate{crts},
	}

	if !info.Wget {
		tlscfg.ClientCAs = pool
		if info.ForceVerify {
			tlscfg.ClientAuth = tls.RequireAndVerifyClientCert
		} else {
			tlscfg.ClientAuth = tls.VerifyClientCertIfGiven
		}
	} else {
		tlscfg.InsecureSkipVerify = true
	}
	return tlscfg, nil
}
