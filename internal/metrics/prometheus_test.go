package metrics

import (
	"sync"
	"testing"

	"github.com/kong/koko/internal/log"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestPrometheusCounter(t *testing.T) {
	client := newPrometheusClient(log.Logger)

	// Normally a prometheus counter is initialized in an init function.
	// We are making sure counters registered the first time is thread safe when done on demand.
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.Count("test_count", 1, Tag{Key: "service", Value: "test"})
		}()
	}
	wg.Wait()

	collector, err := client.getCollector(counterCollector, "test_count", Tag{Key: "service", Value: "test"})
	require.Nil(t, err)

	counterVec, ok := collector.(*prometheus.CounterVec)
	require.True(t, ok)

	counter, err := counterVec.GetMetricWith(prometheus.Labels{"service": "test"})
	require.Nil(t, err)

	m := &dto.Metric{}
	err = counter.Write(m)
	require.Nil(t, err)
	require.Equal(t, float64(5), *m.Counter.Value)
}

func TestPrometheusGuage(t *testing.T) {
	client := newPrometheusClient(log.Logger)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()
			client.Gauge("test_gauge", float64(v), Tag{Key: "service", Value: "test"})
		}(i)
	}
	wg.Wait()

	client.Gauge("test_gauge", float64(10), Tag{Key: "service", Value: "test"})

	collector, err := client.getCollector(gaugeCollector, "test_gauge", Tag{Key: "service", Value: "test"})
	require.Nil(t, err)

	gaugeVec, ok := collector.(*prometheus.GaugeVec)
	require.True(t, ok)

	gauge, err := gaugeVec.GetMetricWith(prometheus.Labels{"service": "test"})
	require.Nil(t, err)

	m := &dto.Metric{}
	err = gauge.Write(m)
	require.Nil(t, err)
	require.Equal(t, float64(10), m.Gauge.GetValue())
}

func TestPrometheusGuageAdd(t *testing.T) {
	client := newPrometheusClient(log.Logger)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()
			client.GaugeAdd("test_gauge_add", float64(v), Tag{Key: "service", Value: "test_add"})
		}(i)
	}
	wg.Wait()

	client.GaugeAdd("test_gauge_add", float64(10), Tag{Key: "service", Value: "test_add"})

	collector, err := client.getCollector(gaugeCollector, "test_gauge_add", Tag{Key: "service", Value: "test_add"})
	require.Nil(t, err)

	gaugeVec, ok := collector.(*prometheus.GaugeVec)
	require.True(t, ok)

	gauge, err := gaugeVec.GetMetricWith(prometheus.Labels{"service": "test_add"})
	require.Nil(t, err)

	m := &dto.Metric{}
	err = gauge.Write(m)
	require.Nil(t, err)
	require.Equal(t, float64(10), m.Gauge.GetValue())
}

func TestPrometheusHistogram(t *testing.T) {
	client := newPrometheusClient(log.Logger)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()
			client.Histogram("test_histogram", float64(v), Tag{Key: "service", Value: "test"})
		}(i)
	}
	wg.Wait()

	collector, err := client.getCollector(histogramCollector, "test_histogram", Tag{Key: "service", Value: "test"})
	require.Nil(t, err)

	histogramVec, ok := collector.(*prometheus.HistogramVec)
	require.True(t, ok)

	observer, err := histogramVec.GetMetricWith(prometheus.Labels{"service": "test"})
	require.Nil(t, err)

	histogram, _ := observer.(prometheus.Histogram)
	m := &dto.Metric{}
	err = histogram.Write(m)
	require.Nil(t, err)
	require.Equal(t, float64(10), *m.Histogram.SampleSum)
}
