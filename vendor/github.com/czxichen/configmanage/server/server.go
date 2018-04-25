package server

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	http "github.com/valyala/fasthttp"
)

type logger struct{}

func (logger) Printf(format string, args ...interface{}) {
	log.Printf(format, args)
}

var (
	fileHandler http.RequestHandler
)

func Server(cmd *cobra.Command, args []string) error {
	if cpath != "" {
		readconfig(cpath, &Cfg)
	}

	Cfg.Download = strings.Replace(Cfg.Download, "\\", "/", -1)
	if !strings.HasSuffix(Cfg.Download, "/") {
		Cfg.Download += "/"
	}

	File, err := os.Create(Cfg.Logname)
	if err != nil {
		log.Printf("创建日志文件失败:%s\n", err.Error())
		return nil
	}
	log.SetOutput(File)

	Templatedir = Cfg.Download + "template/"
	Parseconfig()

	go Notify(Templatedir, 10, func() {
		Parseconfig()
	})

	fs := &http.FS{Root: Cfg.Download, AcceptByteRange: true}
	fileHandler = fs.NewRequestHandler()
	err = listen(Cfg.Proto, Cfg.IP, Cfg.CrtPath, Cfg.Keypath, Cfg.Logname)
	log.Printf("监听服务失败:%s\n", err.Error())
	return nil
}

func listen(proto, ip, crt, key, logname string) (err error) {
	s := &http.Server{Handler: Router, Logger: logger{}}
	switch proto {
	case "http":
		err = s.ListenAndServe(ip)
	case "https":
		err = s.ListenAndServeTLS(ip, crt, key)
	default:
		err = errors.New("Not found proto")
	}
	return
}

func Router(ctx *http.RequestCtx) {
	log.Printf("# %s from %s;RemoteAddr:%s --> LocallAddr:%s\n", ctx.Method(), ctx.URI().FullURI(),
		ctx.RemoteAddr(), ctx.LocalAddr())
	ctx.Response.Header.Set("Server", "work-stacks")
	switch string(ctx.Path()) {
	case "/":
		ctx.WriteString("That's is ok")
	case "/download":
		download(ctx)
	case "/serverpackage":
		fileserver("template/"+serverpkg, ctx)
	case "/configtemplate":
		fileserver("configtemp.zip", ctx)
	case "/getvalues":
		getconfigvalue(ctx)
	case "/checkconfig":
		checkconfig(ctx)
	default:
		ctx.NotFound()
	}
}

func getconfigvalue(ctx *http.RequestCtx) {
	key := string(ctx.FormValue("key"))
	if key == "" {
		l := strings.Split(ctx.RemoteAddr().String(), ":")
		if len(l) == 2 {
			key = l[0]
		} else {
			ctx.Logger().Printf("can't get key's value\n")
			ctx.Error("can't get key's value", http.StatusNotFound)
			return
		}
	}
	v, ok := getvalue(key)
	if !ok {
		ctx.Logger().Printf("Not found %s valid variables\n", key)
		ctx.Error("Not found valid variables", http.StatusNotFound)
		return
	}
	rel := serverInfo{Relation: Pathrelation, Variable: v}
	buf := http.AcquireByteBuffer()
	err := gob.NewEncoder(buf).Encode(rel)

	if err != nil {
		ctx.Logger().Printf("gob encode faild:%s\n", err)
		ctx.Error("gob encode faild", http.StatusNotFound)
		return
	}
	ctx.Write(buf.B)
	http.ReleaseByteBuffer(buf)
}

func checkconfig(ctx *http.RequestCtx) {
	for key, value := range Pathrelation {
		fmt.Fprintln(ctx, key, value)
	}
	fmt.Fprintf(ctx, "--------------------------------------------------------------\r\n")
	for key, value := range Variables {
		fmt.Fprintln(ctx, key, value)
	}
}

func download(ctx *http.RequestCtx) {
	path := string(ctx.FormValue("file"))
	if len(path) <= 0 {
		ctx.Error("Request error", http.StatusNotFound)
		return
	}

	fileserver(path, ctx)
}

func fileserver(path string, ctx *http.RequestCtx) {
	//	ctx.Response.Header.Set("Content-Disposition", "attachment; filename="+path)
	ctx.Response.Header.Set("filename", filepath.Base(path))
	ctx.Request.SetRequestURI(path)
	fileHandler(ctx)
}
