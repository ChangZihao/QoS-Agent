package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"math/rand"
)

//Define a struct for you collector that contains pointers
//to prometheus descriptors for each metric you wish to expose.
//Note you can also include fields of other types if they provide utility
//but we just won't be exposing them as metrics.
type rdtCollector struct {
	llcMetric *prometheus.Desc
	//barMetric *prometheus.Desc
}

//You must create a constructor for you collector that
//initializes every descriptor and returns a pointer to the collector
func NewrdtCollector() *rdtCollector {
	return &rdtCollector{
		llcMetric: prometheus.NewDesc("llc_metric",
			"Shows llc metrics measured by pqos", []string{"group"}, nil,
		),
		//barMetric: prometheus.NewDesc("bar_metric",
		//	"Shows whether a bar has occurred in our cluster",
		//	nil, nil,
		//),
	}
}

//Each and every collector must implement the Describe function.
//It essentially writes all descriptors to the prometheus desc channel.
func (collector *rdtCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.llcMetric
	//ch <- collector.barMetric
}

//Collect implements required collect function for all promehteus collectors
func (collector *rdtCollector) Collect(ch chan<- prometheus.Metric) {

	//Implement logic here to determine proper metric value to return to prometheus
	//for each descriptor or call other functions that do so.
	metricValue := rand.Float64()

	//Write latest value for each metric in the prometheus metric channel.
	//Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
	ch <- prometheus.MustNewConstMetric(collector.llcMetric, prometheus.GaugeValue, metricValue, "/czh")
	//ch <- prometheus.MustNewConstMetric(collector.barMetric, prometheus.CounterValue, metricValue)

}
