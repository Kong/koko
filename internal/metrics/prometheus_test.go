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

	counterVec, ok := client.getCollector("test_count").(*prometheus.CounterVec)
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

	gaugeVec, ok := client.getCollector("test_gauge").(*prometheus.GaugeVec)
	require.True(t, ok)

	gauge, err := gaugeVec.GetMetricWith(prometheus.Labels{"service": "test"})
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

	histogramVec, ok := client.getCollector("test_histogram").(*prometheus.HistogramVec)
	require.True(t, ok)

	observer, err := histogramVec.GetMetricWith(prometheus.Labels{"service": "test"})
	require.Nil(t, err)

	histogram, _ := observer.(prometheus.Histogram)
	m := &dto.Metric{}
	err = histogram.Write(m)
	require.Nil(t, err)
	require.Equal(t, float64(10), *m.Histogram.SampleSum)
}
