package cli

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

var registry = prometheus.NewRegistry()

// MonitorHandler monitor Handler
func MonitorHandler(gather prometheus.Gatherer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mfs, err := gather.Gather()
		if err != nil {
			http.Error(w, "An error has occurred during metrics collection:\n\n"+err.Error(), http.StatusInternalServerError)
			return
		}

		contentType := expfmt.Negotiate(req.Header)
		buf := bytes.NewBuffer(nil)
		writer, encoding := decorateWriter(req, buf)
		enc := expfmt.NewEncoder(writer, contentType)
		var lastErr error
		for _, mf := range mfs {
			if err := enc.Encode(mf); err != nil {
				lastErr = err
				http.Error(w, "An error has occurred during metrics encoding:\n\n"+err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if closer, ok := writer.(io.Closer); ok {
			closer.Close()
		}
		if lastErr != nil && buf.Len() == 0 {
			http.Error(w, "No metrics encoded, last error:\n\n"+lastErr.Error(), http.StatusInternalServerError)
			return
		}
		header := w.Header()
		header.Set("Content-Type", string(contentType))
		header.Set("Content-Length", fmt.Sprint(buf.Len()))
		if encoding != "" {
			header.Set("Content-Encoding", encoding)
		}
		w.Write(buf.Bytes())
	})
}

// MonitorInstrument 初始化,func() 执行设置当前的运行状态,Gather获取metric项
func MonitorInstrument(processNames []string) (func(), prometheus.Gatherer) {
	var baseOpts = prometheus.GaugeOpts{
		Namespace: "system",
		Subsystem: "usage",
	}
	cpuOpts := baseOpts
	cpuOpts.Name = "cpu"
	cpuOpts.Help = "CPU 使用率,-1 采集错误"
	cpuGauge := prometheus.NewGauge(cpuOpts)
	registry.MustRegister(cpuGauge)

	memOpts := baseOpts
	memOpts.Name = "mem"
	memOpts.Help = "Mem 使用率,-1 采集错误"
	memGauge := prometheus.NewGauge(memOpts)
	registry.MustRegister(memGauge)

	diskOpts := baseOpts
	diskOpts.Name = "disk"
	diskOpts.Help = "Disk 使用率"
	diskOpts.ConstLabels = make(prometheus.Labels)
	var diskMap = make(map[string]prometheus.Gauge)
	for _, mount := range getDiskMount() {
		diskOpts.ConstLabels["mount"] = mount
		diskGauge := prometheus.NewGauge(diskOpts)
		registry.MustRegister(diskGauge)
		diskMap[mount] = diskGauge
	}

	netOpts := baseOpts
	netOpts.Name = "net"
	netOpts.Help = "Net 使用收发详情"
	netOpts.ConstLabels = make(prometheus.Labels)
	var netMap = make(map[string]prometheus.Gauge)
	for _, name := range getInterfaces() {
		netOpts.ConstLabels["card"] = name
		netOpts.ConstLabels["direction"] = "in"
		netInGauge := prometheus.NewGauge(netOpts)
		registry.MustRegister(netInGauge)
		netMap[name+"_in"] = netInGauge

		netOpts.ConstLabels["direction"] = "out"
		netOutGauge := prometheus.NewGauge(netOpts)
		registry.MustRegister(netOutGauge)
		netMap[name+"_out"] = netOutGauge
	}

	var processMap = make(map[string]prometheus.Gauge)
	if len(processNames) > 0 {
		processOpts := prometheus.GaugeOpts{Namespace: "process", Subsystem: "status", Name: "running",
			Help: "process 存活状态,0 未运行 1 运行中", ConstLabels: make(prometheus.Labels)}
		for _, processName := range processNames {
			processOpts.ConstLabels["process"] = processName
			processGauge := prometheus.NewGauge(processOpts)
			registry.MustRegister(processGauge)
			processMap[processName] = processGauge
		}
	}

	return func() {
		setCPUValue(cpuGauge)
		setMemValue(memGauge)
		setDiskValue(diskMap)
		setInterfaceValue(netMap)
		setProcessStatus(processMap)
	}, registry
}

func getDiskMount() []string {
	pathInfo, err := disk.Partitions(false)
	if err != nil {
		panic(err)
	}
	mounts := make([]string, 0, len(pathInfo))
	for _, path := range pathInfo {
		mounts = append(mounts, path.Mountpoint)
	}
	return mounts
}

func getInterfaces() []string {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	interfaceArray := make([]string, 0, len(interfaces))
	for _, intf := range interfaces {
		interfaceArray = append(interfaceArray, intf.Name)
	}
	return interfaceArray
}

// setCPUValue 设置CPU每秒总使用率,-1表示采集错误
func setCPUValue(gauge prometheus.Gauge) {
	cpuValue, err := cpu.Percent(time.Second, false)
	if err != nil {
		gauge.Set(-1)
	} else {
		gauge.Set(trunc(cpuValue[0]))
	}
}

// setMemValue 设置内存的使用率,-1表示采集错误
func setMemValue(gauge prometheus.Gauge) {
	stat, err := mem.VirtualMemory()
	if err != nil {
		gauge.Set(-1)
	} else {
		gauge.Set(trunc(stat.UsedPercent))
	}
}

// setDiskValue 设置指定挂载点的磁盘使用率
func setDiskValue(diskStatusMap map[string]prometheus.Gauge) {
	for path, gauge := range diskStatusMap {
		stat, err := disk.Usage(path)
		if err == nil {
			gauge.Set(trunc(stat.UsedPercent))
		} else {
			gauge.Set(-1)
			fmt.Printf("Read disk usage error:%s\n", err.Error())
		}
	}
}

// setInterfaceValue 设置网络收发数据
func setInterfaceValue(netStatusMap map[string]prometheus.Gauge) {
	for interfaceName := range netStatusMap {
		netStatusMap[interfaceName].Set(-1)
	}

	ioCounters, err := net.IOCounters(true)
	if err == nil {
		for _, status := range ioCounters {
			if gauge, ok := netStatusMap[status.Name+"_in"]; ok {
				gauge.Set(float64(status.BytesRecv))
			}
			if gauge, ok := netStatusMap[status.Name+"_out"]; ok {
				gauge.Set(float64(status.BytesSent))
			}
		}
	} else {
		fmt.Printf("GetInterfaceUsage error:%s\n", err.Error())
	}
}

func setProcessStatus(isRunningMap map[string]prometheus.Gauge) {
	if len(isRunningMap) <= 0 {
		return
	}
	for processName := range isRunningMap {
		isRunningMap[processName].Set(0)
	}
	ps, err := process.Processes()
	if err == nil {
		for _, p := range ps {
			if processName, err := p.Name(); err == nil {
				if _, ok := isRunningMap[processName]; ok {
					isRunningMap[processName].Set(1)
				}
			}
		}
	} else {
		fmt.Printf("GetProcessList error:%s\n", err.Error())
	}
}

func trunc(f float64) float64 {
	return math.Trunc((f+0.5/1000)*1000) / 1000
}

func decorateWriter(request *http.Request, writer io.Writer) (io.Writer, string) {
	header := request.Header.Get("Accept-Encoding")
	parts := strings.Split(header, ",")
	for _, part := range parts {
		part := strings.TrimSpace(part)
		if part == "gzip" || strings.HasPrefix(part, "gzip;") {
			return gzip.NewWriter(writer), "gzip"
		}
	}
	return writer, ""
}
