package metrics

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type prometheusLog func(msg string, fields ...zapcore.Field)

func (l prometheusLog) Println(v ...interface{}) {
	l(fmt.Sprintf("%v", v))
}

type prometheusClient struct {
	lock       sync.RWMutex
	collectors map[string]prometheus.Collector
	registry   *prometheus.Registry
}

func newPrometheusClient() *prometheusClient {
	return &prometheusClient{
		lock:       sync.RWMutex{},
		collectors: map[string]prometheus.Collector{},
		registry:   prometheus.NewRegistry(),
	}
}

func (c *prometheusClient) getCollector(name string) prometheus.Collector {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.collectors[name]
}

func (c *prometheusClient) Gauge(name string, value float64, tags ...Tag) error {
	c.lock.RLock()
	if gauge, ok := c.collectors[name].(prometheus.GaugeVec); ok {
		gauge.With(convertToLabels(tags...)).Add(value)
		c.lock.RUnlock()
		return nil
	}
	c.lock.RUnlock()
	c.lock.Lock()
	defer c.lock.Unlock()

	opts := prometheus.GaugeOpts{
		Name:      name,
		Namespace: metricNamespace,
	}

	collector, err := c.register(prometheus.NewGaugeVec(opts,
		convertTagsToNames(tags...)))
	if err != nil {
		return err
	}

	gauge, _ := collector.(*prometheus.GaugeVec)
	gauge.With(convertToLabels(tags...)).Add(value)
	c.collectors[name] = gauge
	return nil
}

func (c *prometheusClient) Count(name string, value int64, tags ...Tag) error {
	c.lock.RLock()
	if counter, ok := c.collectors[name].(prometheus.CounterVec); ok {
		counter.With(convertToLabels(tags...)).Add(float64(value))
		c.lock.RUnlock()
		return nil
	}
	c.lock.RUnlock()
	c.lock.Lock()
	defer c.lock.Unlock()

	opts := prometheus.CounterOpts{
		Name:      name,
		Namespace: metricNamespace,
	}

	collector, err := c.register(prometheus.NewCounterVec(opts,
		convertTagsToNames(tags...)))
	if err != nil {
		return err
	}

	counter, _ := collector.(*prometheus.CounterVec)
	counter.With(convertToLabels(tags...)).Add(float64(value))
	c.collectors[name] = counter
	return nil
}

func (c *prometheusClient) Histogram(name string, value float64, tags ...Tag) error {
	c.lock.RLock()
	if histogram, ok := c.collectors[name].(prometheus.HistogramVec); ok {
		histogram.With(convertToLabels(tags...)).Observe(value)
		c.lock.RUnlock()
		return nil
	}
	c.lock.RUnlock()
	c.lock.Lock()
	defer c.lock.Unlock()

	opts := prometheus.HistogramOpts{
		Name:      name,
		Namespace: metricNamespace,
	}

	collector, err := c.register(prometheus.NewHistogramVec(opts,
		convertTagsToNames(tags...)))
	if err != nil {
		return err
	}

	histogram, _ := collector.(*prometheus.HistogramVec)
	histogram.With(convertToLabels(tags...)).Observe(value)
	c.collectors[name] = histogram
	return nil
}

func (c *prometheusClient) CreateHandler(log *zap.Logger) http.Handler {
	handler := promhttp.HandlerFor(c.registry, promhttp.HandlerOpts{
		Registry: c.registry,
		ErrorLog: prometheusLog(log.With(zap.String("server", "prometheus")).Info),
	})
	mux := http.NewServeMux()
	mux.Handle("/metrics", handler)
	return mux
}

func (c *prometheusClient) Close() {}

func (c *prometheusClient) register(collector prometheus.Collector) (prometheus.Collector, error) {
	if err := c.registry.Register(collector); err != nil {
		if arerr, ok := err.(prometheus.AlreadyRegisteredError); ok {
			return arerr.ExistingCollector, nil
		}
		return nil, err
	}
	return collector, nil
}

func convertTagsToNames(tags ...Tag) []string {
	names := make([]string, 0, len(tags))
	for _, tag := range tags {
		names = append(names, tag.Name)
	}
	return names
}

func convertToLabels(tags ...Tag) prometheus.Labels {
	labels := make(prometheus.Labels, len(tags))
	for _, tag := range tags {
		labels[tag.Name] = tag.Value
	}
	return labels
}
