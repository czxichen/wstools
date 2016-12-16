package command

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/valyala/fasthttp"
)

type fileServerConfig struct {
	Addr        string
	Host        string
	SavePath    string
	Dir         string
	IndexFile   bool
	AllowUpload bool
}

var (
	fileServerCFG fileServerConfig
	FileServer    = &Command{
		UsageLine: "fileserver -l 192.168.0.2:8080",
		Run:       fileServer,
		Short:     "用过http协议传输共享文件",
		Long: `当共享文件的时候可以直接在浏览器Get访问http://host:port/dir/filepath如果想上传文件,则需要使用
POST方法访问http://host:port/,当要上传的文件在服务的存在的时候,则POST到http://host:port/?cover=true地
址上.
	wstools fileserver -d command  -i -a
	wstools fileserver -w http://localhost:8080/command/ssh.go -s="../"
`,
	}
)

func init() {
	FileServer.Flag.StringVar(&fileServerCFG.Host, "w", "", `-w="http://localhost:8080/test.txt" 用来指定要下载文件的URL,只能和-s结合使用`)
	FileServer.Flag.StringVar(&fileServerCFG.SavePath, "s", "", `-s="test.txt" 指定文件保存的路径,当-w不为空的时候有效,为空则保存当前目录`)
	FileServer.Flag.StringVar(&fileServerCFG.Addr, "l", ":8080", `-l="192.168.0.2:8080" 指定监听的地址和端口`)
	FileServer.Flag.StringVar(&fileServerCFG.Dir, "d", "./", `-d ="dirs" 指定要共享的目录`)
	FileServer.Flag.BoolVar(&fileServerCFG.IndexFile, "i", false, `-i=true 是否启用文件索引,允许查看目录所有文件`)
	FileServer.Flag.BoolVar(&fileServerCFG.AllowUpload, "a", false, `-a=true 开启文件上传,默认上传到共享目录`)
}

func fileServer(cmd *Command, args []string) bool {
	if fileServerCFG.Host != "" {
		err := wget(fileServerCFG.Host)
		if err != nil {
			log.Printf("Download file error,%s\n", err.Error())
		}
		return true
	}
	fileServerCFG.Dir = strings.Replace(fileServerCFG.Dir, "\\", "/", -1)
	if !strings.HasSuffix(fileServerCFG.Dir, "/") {
		fileServerCFG.Dir += "/"
	}
	fsHandler = fasthttp.FSHandler(fileServerCFG.Dir, 0)
	err := fasthttp.ListenAndServe(fileServerCFG.Addr, router)
	if err != nil {
		log.Println(err)
	}
	return true
}

var fsHandler fasthttp.RequestHandler

func router(ctx *fasthttp.RequestCtx) {
	log.Printf("Remoteaddr:%s Uri:%s\n", ctx.RemoteAddr().String(), ctx.RequestURI())
	ctx.Response.Header.Set("server", "work-stacks")
	switch string(ctx.Method()) {
	case "GET":
		if !fileServerCFG.IndexFile {
			path := fileServerCFG.Dir + strings.TrimPrefix(string(ctx.URI().Path()), "/")
			stat, err := os.Lstat(path)
			if err != nil {
				ctx.Error("Access not allow", 403)
				return
			}
			if stat.IsDir() {
				ctx.Error("Access not allow", 403)
				return
			}
		}
		fsHandler(ctx)
	case "POST":
		if !fileServerCFG.AllowUpload {
			ctx.Error("Access not allow", 403)
			return
		}
		form, err := ctx.MultipartForm()
		if err != nil {
			ctx.Error("Get multipart faild", 500)
			return
		}

		for _, headers := range form.File {
			file := fileServerCFG.Dir + headers[0].Filename
			_, err = os.Lstat(file)
			if !os.IsNotExist(err) {
				if string(ctx.FormValue("cover")) != "true" {
					ctx.Error("File is exist.", 403)
					return
				}
			}
			err = fasthttp.SaveMultipartFile(headers[0], file)
			if err != nil {
				ctx.Error(fmt.Sprintf("Save file %s faild", headers[0].Filename), 500)
				return
			}
		}
	}
}

func wget(url string) error {
	var maxRedirectsCount = 10
	client := fasthttp.Client{TLSConfig: &tls.Config{InsecureSkipVerify: true}}
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	req.Header.Set("User-Agent", "wstools")
	redirectsCount := 0
	var err error
	for {
		req.SetRequestURI(url)
		err = client.Do(req, resp)
		if err != nil {
			break
		}

		if resp.StatusCode() != 301 && resp.StatusCode() != 302 && resp.StatusCode() != 303 {
			break
		}

		if redirectsCount > maxRedirectsCount {
			err = errors.New("Too many redirects detected when doing the request")
			break
		}

		location := resp.Header.Peek("Location")
		if len(location) == 0 {
			err = errors.New("missing Location header for http redirect")
			break
		}
		url = getRedirectURL(url, location)
	}
	if err != nil {
		return err
	}

	filename := string(resp.Header.Peek("Content-Disposition"))
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
		list := strings.Split(url, "/")
		filename = list[len(list)-1]
	}

	if fileServerCFG.SavePath != "" {
		fileServerCFG.SavePath = strings.Replace(fileServerCFG.SavePath, "\\", "/", -1)
		if !strings.HasSuffix(fileServerCFG.SavePath, "/") {
			fileServerCFG.SavePath += "/"
		}
		filename = fileServerCFG.SavePath + filename
	}

	File, err := os.Create(filename)
	if err != nil {
		return err
	}
	err = resp.BodyWriteTo(File)
	File.Close()
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(resp)
	return err
}

func getRedirectURL(baseURL string, location []byte) string {
	u := fasthttp.AcquireURI()
	u.Update(baseURL)
	u.UpdateBytes(location)
	redirectURL := u.String()
	fasthttp.ReleaseURI(u)
	return redirectURL
}
