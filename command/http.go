package command

import (
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/lucas-clemente/quic-go/h2quic"
)

var FileServer = &Command{
	UsageLine: "http -l 127.0.0.1:8080",
	Run:       fileServer,
	Short:     "用过http协议传输共享文件",
	Long: `当共享文件的时候可以直接在浏览器Get访问http://host:port/dir/filepath
可以使用Quic协议通过UDP共享

	wstools http -d dir -index
	wstools http -quic -onlyquic -crt root.crt -key root.key
	wstools http -w http://127.0.0.1:8080/dir/ssh.go -s="../"
	wstools http -w https://127.0.0.1:8080/cfg.json -crt agent.crt -key agent.key -quic
`,
}

var fileServerCFG = new(fileServerConfig)

func init() {
	FileServer.Flag.StringVar(&fileServerCFG.listen, "l", ":8080", `-l="192.168.0.2:8080" 指定监听的地址和端口`)
	FileServer.Flag.StringVar(&fileServerCFG.requst, "w", "", `-w="http://localhost:8080/test.txt" 用来指定要下载文件的URL,只能和-s结合使用`)
	FileServer.Flag.StringVar(&fileServerCFG.savepath, "s", "", `-s="test.txt" 指定文件保存的路径,当-w不为空的时候有效,为空则保存当前目录`)
	FileServer.Flag.StringVar(&fileServerCFG.sharedir, "d", "./", `-d ="dirs" 指定要共享的目录`)
	FileServer.Flag.StringVar(&fileServerCFG.crtpath, "crt", "", "-crt client.crt|-crt root.crt 指定证书公钥")
	FileServer.Flag.StringVar(&fileServerCFG.keypath, "key", "", "-key client.key|-key root.key 指定证书私钥")
	FileServer.Flag.BoolVar(&fileServerCFG.allowindex, "index", false, `-index 是否启用文件索引,允许查看目录所有文件`)
	FileServer.Flag.BoolVar(&fileServerCFG.allowquic, "quic", false, `-quic 启用quic协议,必须指定-crt和-key参数`)
	FileServer.Flag.BoolVar(&fileServerCFG.onlyquic, "onlyquic", false, `-onlyquic 只使用quic协议,不启用tcp模式`)
}

func fileServer(cmd *Command, args []string) bool {
	if fileServerCFG.requst != "" {
		err := fileServerCFG.wget()
		if err != nil {
			log.Printf("Requst error:%s\n", err.Error())
		}
		return !os.IsNotExist(err)
	}

	var dir = strings.Replace(fileServerCFG.sharedir, "\\", "/", -1)
	if dir[len(dir)-1] == '/' {
		dir = dir[:len(dir)-1]
	}
	info, err := os.Lstat(dir)
	if err != nil || !info.IsDir() {
		var errinfo string
		if err != nil {
			errinfo = err.Error()
		} else {
			errinfo = "path is not a directory"
		}
		log.Fatalf("Share directory faild:%s\n", errinfo)
	}

	fileServerCFG.handler = http.FileServer(http.Dir(dir))
	fileServerCFG.sharedir = dir

	if fileServerCFG.crtpath != "" && fileServerCFG.keypath != "" {
		var httpserver func(string, string, string, http.Handler) error
		if fileServerCFG.allowquic {
			if fileServerCFG.onlyquic {
				httpserver = h2quic.ListenAndServeQUIC
			} else {
				httpserver = h2quic.ListenAndServe
			}
		} else {
			httpserver = http.ListenAndServeTLS
		}
		err = httpserver(fileServerCFG.listen, fileServerCFG.crtpath, fileServerCFG.keypath, fileServerCFG)
	} else {
		err = http.ListenAndServe(fileServerCFG.listen, fileServerCFG)
	}
	if err != nil {
		log.Printf("Listen and server error:%s\n", err.Error())
	}
	return !os.IsNotExist(err)
}

type fileServerConfig struct {
	listen     string
	requst     string
	savepath   string
	sharedir   string
	crtpath    string
	keypath    string
	allowquic  bool
	onlyquic   bool
	allowindex bool
	handler    http.Handler
}

func (fs *fileServerConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("From:%s\tRequest:%s", r.RemoteAddr, r.RequestURI)
	r.Header.Set("server", "wstools")
	if !fs.allowindex {
		info, err := os.Lstat(fs.sharedir + r.URL.Path)
		if err != nil {
			http.Error(w, "Not Found", 404)
			return
		}
		if info.IsDir() {
			http.Error(w, "Not Allow", 403)
			return
		}
	}
	if r.Method == "GET" {
		fs.handler.ServeHTTP(w, r)
	} else {
		http.Error(w, "Server unsupport this method", 500)
	}
}

func (fc *fileServerConfig) wget() error {
	req, err := http.NewRequest("GET", fc.requst, nil)
	if err != nil {
		return err
	}
	var client = &http.Client{}
	if strings.HasPrefix(fc.requst, "https") || fc.allowquic {
		var tlscfg = &tls.Config{InsecureSkipVerify: true}
		if fc.crtpath != "" && fc.keypath != "" {
			crt, err := tls.LoadX509KeyPair(fc.crtpath, fc.keypath)
			if err != nil {
				return err
			}
			tlscfg.Certificates = []tls.Certificate{crt}
		}
		if !fc.allowquic {
			client.Transport = &http.Transport{TLSClientConfig: tlscfg}
		} else {
			client.Transport = &h2quic.QuicRoundTripper{TLSClientConfig: tlscfg}
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	var filename = string(resp.Header.Get("Content-Disposition"))
	if len(filename) > 0 {
		for _, name := range strings.Split(filename, ";") {
			if strings.Contains(name, "filename") {
				list := strings.Split(name, "=")
				if len(list) == 2 {
					filename = strings.TrimSpace(list[1])
				}
				break
			}
		}
	}

	if filename == "" {
		list := strings.Split(req.URL.Path, "/")
		filename = list[len(list)-1]
	}

	if fileServerCFG.savepath != "" {
		fileServerCFG.savepath = strings.Replace(fileServerCFG.savepath, "\\", "/", -1)
		if strings.HasSuffix(fileServerCFG.savepath, "/") {
			filename = fileServerCFG.savepath + filename
		} else {
			filename = fileServerCFG.savepath
		}
	}

	File, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer File.Close()
	_, err = io.Copy(File, resp.Body)
	return err
}
