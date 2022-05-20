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

type collectorType int

const (
	counterCollector collectorType = iota
	gaugeCollector
	histogramCollector
)

type prometheusLog func(msg string, fields ...zapcore.Field)

func (l prometheusLog) Println(v ...interface{}) {
	l(fmt.Sprintf("%v", v))
}

type prometheusClient struct {
	log        *zap.Logger
	collectors sync.Map
	registry   *prometheus.Registry
}

func newPrometheusClient(logger *zap.Logger) *prometheusClient {
	return &prometheusClient{
		log:      logger,
		registry: prometheus.NewRegistry(),
	}
}

func (c *prometheusClient) getCollector(collectorType collectorType,
	name string, tags ...Tag,
) (prometheus.Collector, error) {
	col, ok := c.collectors.Load(name)
	if ok {
		if collector, ok := col.(prometheus.Collector); ok {
			return collector, nil
		}
		return nil, fmt.Errorf("metric name '%s' is not a collector", name)
	}

	var collector prometheus.Collector
	switch collectorType {
	case counterCollector:
		collector = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: name,
		}, convertTagsToNames(tags...))
	case gaugeCollector:
		collector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: name,
		}, convertTagsToNames(tags...))
	case histogramCollector:
		collector = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: name,
		}, convertTagsToNames(tags...))
	default:
		panic("unsupported prometheus collector")
	}

	collector, err := c.register(collector)
	if err != nil {
		return nil, fmt.Errorf("failed to register: %w", err)
	}
	c.collectors.Store(name, collector)
	return collector, nil
}

func (c *prometheusClient) Gauge(name string, value float64, tags ...Tag) {
	collector, err := c.getCollector(gaugeCollector, name, tags...)
	if err != nil {
		c.log.With(zap.Error(err)).Error("failed to update gauge")
		return
	}
	if gauge, ok := collector.(*prometheus.GaugeVec); ok {
		gauge.With(convertToLabels(tags...)).Set(value)
		return
	}
	c.log.Error("collector is not a gauge", zap.String("name", name))
}

func (c *prometheusClient) GaugeAdd(name string, value float64, tags ...Tag) {
	collector, err := c.getCollector(gaugeCollector, name, tags...)
	if err != nil {
		c.log.With(zap.Error(err)).Error("failed to update gauge")
		return
	}
	if gauge, ok := collector.(*prometheus.GaugeVec); ok {
		gauge.With(convertToLabels(tags...)).Add(value)
		return
	}
	c.log.Error("collector is not a gauge", zap.String("name", name))
}

func (c *prometheusClient) Count(name string, value int64, tags ...Tag) {
	collector, err := c.getCollector(counterCollector, name, tags...)
	if err != nil {
		c.log.With(zap.Error(err)).Error("failed to update counter")
		return
	}

	if counter, ok := collector.(*prometheus.CounterVec); ok {
		counter.With(convertToLabels(tags...)).Add(float64(value))
		return
	}
	c.log.Error("collector is not a counter", zap.String("name", name))
}

func (c *prometheusClient) Histogram(name string, value float64, tags ...Tag) {
	collector, err := c.getCollector(histogramCollector, name, tags...)
	if err != nil {
		c.log.With(zap.Error(err)).Error("failed to update histogram")
		return
	}

	if histogram, ok := collector.(*prometheus.HistogramVec); ok {
		histogram.With(convertToLabels(tags...)).Observe(value)
		return
	}
	c.log.Error("collector is not a histogram", zap.String("name", name))
}

func (c *prometheusClient) CreateHandler(log *zap.Logger) (http.Handler, error) {
	handler := promhttp.HandlerFor(c.registry, promhttp.HandlerOpts{
		Registry: c.registry,
		ErrorLog: prometheusLog(log.With(zap.String("server", "prometheus")).Error),
	})
	mux := http.NewServeMux()
	mux.Handle("/metrics", handler)
	return mux, nil
}

func (c *prometheusClient) Close() error {
	return nil
}

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
		names = append(names, tag.Key)
	}
	return names
}

func convertToLabels(tags ...Tag) prometheus.Labels {
	labels := make(prometheus.Labels, len(tags))
	for _, tag := range tags {
		labels[tag.Key] = tag.Value
	}
	return labels
}
