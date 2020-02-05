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
	"os/exec"
	"strings"
)

var (
	// Set during go build
	// version   string
	// gitCommit string

	// 命令行参数
	listenAddr       = flag.String("web.listen-port", "9001", "An port to listen on for web interface and telemetry.")
	metricsPath      = flag.String("web.telemetry-path", "/metrics", "A path under which to expose metrics.")
	metricsNamespace = flag.String("metric.namespace", "pqos", "Prometheus metrics namespace, as the prefix of metrics name")
)

func main() {

	flag.Parse()

	metrics := collector.NewpqosCollector()
	registry := prometheus.NewRegistry()
	registry.MustRegister(metrics)

	http.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	http.HandleFunc("/monitor/start", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		pod := r.Form.Get("pod")
		app := r.Form.Get("app")
		pod = "pod" + strings.ReplaceAll(pod, "-", "_")
		if pod != "" {
			if _, ok := collector.MonitorCMD.Load(pod); !ok {
				cmd := podInfo.StarqposMonitor(pod, &collector.PQOSMetrics)
				collector.MonitorCMD.Store(pod, cmd)
				collector.Pod2app.Store(pod, app)
				fmt.Fprintf(w, "Start monitor %s\n", pod)
				log.Infof("Start monitor app %s, pod  %s\n", app, pod)
			} else {
				fmt.Fprintf(w, "Monitor for %s already existed!\n", pod)
				log.Infof("Monitor for app %s, %s already existed!\n", app, pod)
			}
		} else {
			log.Errorln("pod name is nil")
		}
	})

	http.HandleFunc("/monitor/stop", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		pod := r.Form.Get("pod")
		pod = "pod" + strings.ReplaceAll(pod, "-", "_")
		if pod != "" {
			cmd, ok := collector.MonitorCMD.Load(pod)
			if ok {
				podInfo.StopMonitor(cmd.(*exec.Cmd))
				collector.MonitorCMD.Delete(pod)
				collector.Pod2app.Delete(pod)
				fmt.Fprintf(w, "Stop monitor %s\n", pod)
			} else {
				fmt.Fprintf(w, "Fail to find %s\n", pod)
			}
			collector.PQOSMetrics.Delete(pod)
		} else {
			log.Errorln("pod name is nil")
		}
	})

	log.Infof("Starting Server at http://localhost:%s%s", *listenAddr, *metricsPath)
	log.Fatal(http.ListenAndServe(":"+*listenAddr, nil))
}
