package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type catalog struct {
	evaluateEvery             time.Duration
	numCounters               int
	numCounterWithExemplars   int
	numGauges                 int
	numHistograms             int
	numHistogramWithExemplars int
	numNativeHistograms       int
}

var (
	counters         []prometheus.Counter
	gauges           []prometheus.Gauge
	histograms       []prometheus.Histogram
	nativeHistograms []prometheus.Histogram
)

func main() {
	c := new(catalog)
	parseFlags(c, os.Args[1:])
	generateMetricsExemplars(c)

	go func() {
		t := time.NewTicker(c.evaluateEvery)
		defer t.Stop()
		for range t.C {
			// Counters.
			for i := 0; i < c.numCounters+c.numCounterWithExemplars; i++ {
				if i < c.numCounters {
					// Plain counter.
					counters[i].Add(float64(rand.Intn(10)))
					continue
				}
				// Indexes after c.numCounters are with exemplars.
				if adder, ok := counters[i].(prometheus.ExemplarAdder); ok {
					adder.AddWithExemplar(float64(rand.Intn(10)), prometheus.Labels{
						"TraceID":      generateRandomStr(10),
						"job":          "generator",
						"random_label": generateRandomStr(5),
					})
				}
			}
			// Gauges.
			for i := 0; i < c.numGauges; i++ {
				gauges[i].Add(float64(rand.Intn(10)))
			}
			// Histograms.
			for i := 0; i < c.numHistograms+c.numHistogramWithExemplars; i++ {
				if i < c.numHistograms {
					// Plain histogram.
					histograms[i].Observe(float64(rand.Float32()))
					continue
				}
				// Indexes after c.numHistograms are with exemplars.
				if adder, ok := histograms[i].(prometheus.ExemplarObserver); ok {
					adder.ObserveWithExemplar(float64(rand.Intn(10)), prometheus.Labels{
						"TraceID":      generateRandomStr(10),
						"job":          "generator",
						"random_label": generateRandomStr(5),
					})
				}
			}
			// Native high-res histograms.
			for i := 0; i < c.numNativeHistograms; i++ {
				nativeHistograms[i].Observe(float64(rand.Float32()))
			}
		}
	}()

	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			Registry: prometheus.DefaultRegisterer,
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))
	err := http.ListenAndServe(":9001", nil)
	if err != nil {
		panic(err)
	}
}

func generateMetricsExemplars(c *catalog) {
	var (
		counter         []prometheus.Counter
		gauge           []prometheus.Gauge
		histogram       []prometheus.Histogram
		nativeHistogram []prometheus.Histogram
	)
	for i := 0; i < c.numCounters; i++ {
		tmp := prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "metrics_gen",
			Name:      fmt.Sprintf("counter_%d_total", i),
			Help:      fmt.Sprintf("Generated counter num %d.", i),
		})
		counter = append(counter, tmp)
		prometheus.MustRegister(tmp)
	}
	for i := 0; i < c.numCounterWithExemplars; i++ {
		tmp := prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "metrics_exemplars_gen",
			Name:      fmt.Sprintf("counter_%d_total", i),
			Help:      fmt.Sprintf("Generated counter exemplar num %d.", i),
		})
		counter = append(counter, tmp)
		prometheus.MustRegister(tmp)
	}
	for i := 0; i < c.numGauges; i++ {
		tmp := prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "metrics_exemplars_gen",
			Name:      fmt.Sprintf("gauge_%d", i),
			Help:      fmt.Sprintf("Generated gauge num %d.", i),
		})
		gauge = append(gauge, tmp)
		prometheus.MustRegister(tmp)
	}
	for i := 0; i < c.numHistograms; i++ {
		tmp := prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "metrics_gen",
			Name:      fmt.Sprintf("histogram_%d", i),
			Help:      fmt.Sprintf("Generated histogram num %d.", i),
			Buckets:   prometheus.DefBuckets,
		})
		histogram = append(histogram, tmp)
		prometheus.MustRegister(tmp)
	}
	for i := 0; i < c.numHistogramWithExemplars; i++ {
		tmp := prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "metrics_exemplars_gen",
			Name:      fmt.Sprintf("histogram_%d", i),
			Help:      fmt.Sprintf("Generated histogram exemplar num %d.", i),
			Buckets:   prometheus.DefBuckets,
		})
		histogram = append(histogram, tmp)
		prometheus.MustRegister(tmp)
	}
	for i := 0; i < c.numNativeHistograms; i++ {
		tmp := prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:                       "metrics_gen",
			Name:                            fmt.Sprintf("native_histogram_%d", i),
			Help:                            fmt.Sprintf("Generated native histogram num %d", i),
			Buckets:                         prometheus.DefBuckets,
			NativeHistogramBucketFactor:     1.1,
			NativeHistogramMaxBucketNumber:  150,
			NativeHistogramMinResetDuration: time.Hour,
		})
		nativeHistogram = append(nativeHistogram, tmp)
		prometheus.MustRegister(tmp)
	}
	counters = counter
	gauges = gauge
	histograms = histogram
	nativeHistograms = nativeHistogram
}

func parseFlags(conf *catalog, args []string) {
	flag.DurationVar(&conf.evaluateEvery, "evaluate-every", time.Second, "Frequency of evaluation of metrics and exemplar.")
	flag.IntVar(&conf.numCounters, "num-counters", 1, "Number of counters to be generated.")
	flag.IntVar(&conf.numCounterWithExemplars, "num-counters-with-exemplars", 1, "Number of counters to be generated with exemplars.")
	flag.IntVar(&conf.numGauges, "num-gauges", 1, "Number of gauges to be generated.")
	flag.IntVar(&conf.numHistograms, "num-histograms", 1, "Number of histograms to be generated.")
	flag.IntVar(&conf.numHistogramWithExemplars, "num-histograms-with-exemplars", 1, "Number of histograms to be generated with exemplars.")
	flag.IntVar(&conf.numNativeHistograms, "num-native-histograms", 1, "Number of native high-resolution histograms, each with max 150.")
	_ = flag.CommandLine.Parse(args)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func generateRandomStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
