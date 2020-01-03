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
	"sync"
	"time"
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

func starqposMonitor(pod string, m *sync.Map) *exec.Cmd {
	if podPath, isFind := podInfo.GetPod(pod); isFind {
		pids := podInfo.GetPodPids(podPath)
		fmt.Println(pids)
		cmd, stdout := podInfo.Monitor(pids)
		i := 0
		go func() {
			for {
				content := utils.ReadLines(len(pids)+2, 2, stdout)
				data := utils.GetpqosFormat(content)
				if data != nil {
					m.Store(pod, data)
					fmt.Println("test", i)
					i++
				} else {
					m.Store(pod, []float64{0, 0, 0, 0, 0})
					break
				}
			}
			log.Infof("Stop Monitor %s", pod)
		}()
		return cmd
	} else {
		return nil
	}
}

func main() {
	_ = starqposMonitor("podfcf9043f_ccea_4770_acdc_22bb750d6daf", &collector.PQOSMetrics)

	time.Sleep(4 * time.Second)

	//podInfo.StopMonitor(cmd)

	flag.Parse()

	metrics := collector.NewpqosCollector()
	registry := prometheus.NewRegistry()
	registry.MustRegister(metrics)

	http.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	log.Infof("Starting Server at http://localhost:%s%s", *listenAddr, *metricsPath)
	log.Fatal(http.ListenAndServe(":"+*listenAddr, nil))
}
