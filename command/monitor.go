package command

import (
	"bytes"
	"fmt"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/czxichen/wstools/common/cli"
	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	"github.com/spf13/cobra"
)

// Monitor 监控采集节点
var Monitor = &cobra.Command{
	Use: "monitor",
	Example: `	推送方式上传监控项
	-e http://192.168.0.128:9091/metrics -i 5`,
	Short: "Prometheus监控采集节点",
	Long:  "Prometheus监控采集节点,目前仅支持CPU,内存,网络,磁盘,进程存活",
	Run:   monitorRun,
}

type monitorConfig struct {
	JobName  string
	Instance string
	Listen   string
	Prefix   string
	Endpoint string
	Process  []string
	Interval int
}

var monitorCfg monitorConfig

func init() {
	Monitor.PersistentFlags().StringVarP(&monitorCfg.JobName, "job_name", "j", "system", "设置job名称")
	Monitor.PersistentFlags().StringVarP(&monitorCfg.Instance, "instance", "", "", "设置实例名称")
	Monitor.PersistentFlags().StringVarP(&monitorCfg.Listen, "listen", "l", "", "监听地址端口,如果不为空则Endpoint参数不可用")
	Monitor.PersistentFlags().StringVarP(&monitorCfg.Prefix, "prefix", "", "/metric", "指定pull的uri路径,指定Listen时生效")
	Monitor.PersistentFlags().StringVarP(&monitorCfg.Endpoint, "endpoint", "e", "", "PushGateway地址:http://127.0.0.1/metrics")
	Monitor.PersistentFlags().StringArrayVarP(&monitorCfg.Process, "process", "p", nil, "要监控的进程名称")
	Monitor.PersistentFlags().IntVarP(&monitorCfg.Interval, "interval", "i", 60, "采集时间间隔,单位:s")
}

func monitorRun(cmd *cobra.Command, args []string) {
	setValueFunc, gather := cli.MonitorInstrument(monitorCfg.Process)
	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(monitorCfg.Interval))
		defer ticker.Stop()
		for range ticker.C {
			setValueFunc()
		}
	}()

	setValueFunc()
	if monitorCfg.Listen != "" {
		http.Handle(monitorCfg.Prefix, cli.MonitorHandler(gather))
		if err := http.ListenAndServe(monitorCfg.Listen, nil); err != nil {
			fmt.Printf("Listen server error:%s\n", err.Error())
		}
	} else if monitorCfg.Endpoint != "" {
		if monitorCfg.Instance == "" {
			monitorCfg.Instance = getLocalAddress(monitorCfg.Endpoint)
		}
		endpoint := fmt.Sprintf("%s/job/%s/instance/%s", monitorCfg.Endpoint, monitorCfg.JobName, monitorCfg.Instance)
		contentType := mime.FormatMediaType("application/vnd.google.protobuf",
			map[string]string{"encoding": "delimited", "proto": "io.prometheus.client.MetricFamily"})
		ticker := time.NewTicker(time.Second * time.Duration(monitorCfg.Interval))
		defer ticker.Stop()
		for range ticker.C {
			family, err := gather.Gather()
			if err == nil {
				buf := bytes.NewBuffer(nil)
				for _, f := range family {
					pbutil.WriteDelimited(buf, f)
				}
				resp, err := http.Post(endpoint, contentType, buf)
				if err != nil {
					fmt.Printf("send data to %s error:%s\n", endpoint, err.Error())
				} else {
					if resp.StatusCode != http.StatusAccepted {
						fmt.Printf("send to %s data error:%s\n", endpoint, resp.Status)
					}
				}
			}
		}
	} else {
		cli.FatalOutput(1, "参数错误")
	}
}

func getLocalAddress(endpoint string) string {
	hostURL, err := url.Parse(endpoint)
	if err != nil {
		panic(err)
	}
	var host = hostURL.Host
	if strings.Count(host, ":") == 0 {
		host = hostURL.Host + ":80"
	}
	conn, err := net.DialTimeout("tcp", host, 10*time.Second)
	if err == nil {
		host, _, _ = net.SplitHostPort(conn.LocalAddr().String())
		conn.Close()
	} else {
		fmt.Printf("Get host error:%s\n", err.Error())
		host, _ = os.Hostname()
	}
	return host
}
