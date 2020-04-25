package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"net/http"
	"node-exporter/collector"
	"node-exporter/podInfo"
	"node-exporter/utils"
	"os/exec"
	"strings"
	"sync"
)

var (
	// Set during go build
	// version   string
	// gitCommit string

	// 命令行参数
	listenAddr       = flag.String("web.listen-port", "9001", "An port to listen on for web interface and telemetry.")
	metricsPath      = flag.String("web.telemetry-path", "/metrics", "A path under which to expose metrics.")
	metricsNamespace = flag.String("metric.namespace", "pqos", "Prometheus metrics namespace, as the prefix of metrics name")
	masterAddress    = flag.String("master-address", "172.18.13.224:9002", "Qos Master address(ip:port)")

	// 映射pod和pid
	podPids          = sync.Map{}
)

func main() {

	flag.Parse()

	metrics := collector.NewpqosCollector()
	registry := prometheus.NewRegistry()
	registry.MustRegister(metrics)
	llcManager := NewLLCManager()

	// node export 数据展示接口
	http.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	// pod 监控启动接口
	http.HandleFunc("/monitor/start", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		pod := r.Form.Get("pod")
		app := r.Form.Get("app")
		pod = "pod" + strings.ReplaceAll(pod, "-", "_")
		if pod != "" {
			// 判断pod的监控是否已经存在
			if _, ok := collector.MonitorCMD.Load(pod); !ok {
				cmd, pids := collector.StarqposMonitor(pod, &podPids)
				// 保存pod监控的cmd， 便于kill
				collector.MonitorCMD.Store(pod, cmd)
				// 保存pod与app映射
				collector.Pod2app.Store(pod, app)
				// 保存pod与pid映射
				podPids.Store(pod, pids)
				// 初始化pod的llc使用量
				collector.LLCAllocCount.Store(pod, 0)
				fmt.Fprintf(w, "Start monitor %s\n", pod)
				log.Infof("Start monitor app %s, pod  %s\n", app, pod)

				paras := map[string]string{
					"pod": pod,
					"app": app,
					//TODO get host ip auto
					"node": "172.18.13.223",
				}
				url := fmt.Sprintf("http://%s/register", *masterAddress)
				utils.HTTPGet(url, paras)

			} else {
				fmt.Fprintf(w, "Monitor for %s already existed!\n", pod)
				log.Infof("Monitor for app %s, %s already existed!\n", app, pod)
			}
		} else {
			log.Errorln("pod name is nil")
		}
	})

	// pod 监控停止接口
	http.HandleFunc("/monitor/stop", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		pod := r.Form.Get("pod")
		pod = "pod" + strings.ReplaceAll(pod, "-", "_")
		if pod != "" {
			cmd, ok := collector.MonitorCMD.Load(pod)
			if ok {
				// 删除pod的相应信息
				collector.StopMonitor(cmd.(*exec.Cmd))
				collector.MonitorCMD.Delete(pod)
				collector.Pod2app.Delete(pod)
				podPids.Delete(pod)
				fmt.Fprintf(w, "Stop monitor %s\n", pod)
			} else {
				fmt.Fprintf(w, "Fail to find %s\n", pod)
			}
			collector.PQOSMetrics.Delete(pod)
		} else {
			log.Errorln("pod name is nil")
		}
	})

	// pod 资源调控接口
	http.HandleFunc("/control", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		pod := r.Form.Get("pod")
		resourceType := r.Form.Get("resourceType")
		value := r.Form.Get("value")
		log.Infof("Receive resource control: %s : %s : %s", pod, resourceType, value)

		switch resourceType {
		case "cpu_share":
			res := podInfo.SetPodCPUShare(pod, value)
			if res == true {
				w.Write([]byte("succeed"))
			} else {
				w.Write([]byte("failed"))
			}
		case "mem":
			w.Write([]byte("MEM"))
		case "llc":
			if llcManager.AllocLLC(pod, value) {
				w.Write([]byte("succeed"))
			} else {
				w.Write([]byte("failed"))
			}
		}
	})

	log.Infof("Starting Server at http://localhost:%s%s", *listenAddr, *metricsPath)
	log.Fatal(http.ListenAndServe(":"+*listenAddr, nil))
}
