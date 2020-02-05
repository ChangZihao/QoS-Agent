package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

var (
	PQOSMetrics = sync.Map{}
	MonitorCMD  = sync.Map{}
	Pod2app     = sync.Map{}
	metricsList = []string{"ipcMetric", "missedMetric", "llcMetric", "mblMetric", "mbrMetric"}
)

//Define a struct for you collector that contains pointers
//to prometheus descriptors for each metric you wish to expose.
//Note you can also include fields of other types if they provide utility
//but we just won't be exposing them as metrics.
type pqosCollector struct {
	metrics map[string]*prometheus.Desc
	//barMetric *prometheus.Desc
}

//You must create a constructor for you collector that
//initializes every descriptor and returns a pointer to the collector
func NewpqosCollector() *pqosCollector {
	return &pqosCollector{
		metrics: map[string]*prometheus.Desc{
			"ipcMetric":    prometheus.NewDesc("ipc_metric", "Show ipc measured by pqos", []string{"group", "app"}, nil),
			"missedMetric": prometheus.NewDesc("misses_metric", "Show llc misses measured by pqos", []string{"group", "app"}, nil),
			"llcMetric":    prometheus.NewDesc("llc_metric", "Show LLC occupancy measured by pqos", []string{"group", "app"}, nil),
			"mblMetric":    prometheus.NewDesc("mbl_metric", "Show local memory bandwidth measured by pqos", []string{"group", "app"}, nil),
			"mbrMetric":    prometheus.NewDesc("mbr_metric", "Show remote memory bandwidth measured by pqos", []string{"group", "app"}, nil),
		},
	}
}

//Each and every collector must implement the Describe function.
//It essentially writes all descriptors to the prometheus desc channel.
func (collector *pqosCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	for _, m := range collector.metrics {
		ch <- m
	}
}

//Collect implements required collect function for all promehteus collectors
func (collector *pqosCollector) Collect(ch chan<- prometheus.Metric) {

	//Implement logic here to determine proper metric value to return to prometheus
	//for each descriptor or call other functions that do so.
	setData := func(k, v interface{}) bool {
		label := k.(string)
		data := v.([]float64)
		app, _ := Pod2app.Load(label)
		for i, name := range metricsList {
			ch <- prometheus.MustNewConstMetric(collector.metrics[name], prometheus.GaugeValue, data[i], label, app.(string))
		}
		return true
	}
	//Write latest value for each metric in the prometheus metric channel.
	//Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
	PQOSMetrics.Range(setData)
}
