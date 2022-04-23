package metrics

import (
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestPrometheusCounter(t *testing.T) {
	client := newPrometheusClient()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := client.Count("test_count", 1, Tag{Name: "service", Value: "test"})
			require.Nil(t, err)
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
	client := newPrometheusClient()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()
			err := client.Gauge("test_gauge", float64(v), Tag{Name: "service", Value: "test"})
			require.Nil(t, err)
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
	client := newPrometheusClient()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()
			err := client.Histogram("test_histogram", float64(v), Tag{Name: "service", Value: "test"})
			require.Nil(t, err)
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
