package cli

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
)

// HTTPRun HTTPRun
func HTTPRun(httpConfig *HTTPConfig) error {
	var (
		err    error
		tlscfg *tls.Config
	)
	if httpConfig.Wget {
		if strings.HasPrefix(httpConfig.Host, "https://") || httpConfig.Quic {
			tlscfg, err = parseTLS(httpConfig)
			if err != nil {
				return err
			}
		} else {
			if httpConfig.Quic {
				return fmt.Errorf("必须使用https通信")
			}
		}
		return command.Wget(httpConfig.Quic, httpConfig.Host, httpConfig.Save, httpConfig.User, httpConfig.Passwd, tlscfg)
	}

	httpConfig.Dir = filepath.Clean(httpConfig.Dir)

	if httpConfig.Crt != "" {
		if httpConfig.Quic {
			if httpConfig.OnlyQuic {
				err = h2quic.ListenAndServeQUIC(httpConfig.Host, httpConfig.Crt, httpConfig.Key, httpConfig)
			} else {
				err = h2quic.ListenAndServe(httpConfig.Host, httpConfig.Crt, httpConfig.Key, httpConfig)
			}
		} else {
			err = http.ListenAndServeTLS(httpConfig.Host, httpConfig.Crt, httpConfig.Key, httpConfig)
		}
	} else {
		err = http.ListenAndServe(httpConfig.Host, httpConfig)
	}
	return err
}

// HTTPConfig http config
type HTTPConfig struct {
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
	ForceVerify bool // 当前不可用
}

// ServeHTTP ServeHTTP
func (dir HTTPConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func parseTLS(info *HTTPConfig) (*tls.Config, error) {
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
