package v2

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Gauge is a metrics gauge.
// Please consult Prometheus documentation for a detailed understanding.
type Gauge interface {
	// Set sets the Gauge to an arbitrary value.
	Set(float64, ...Label)
	// Inc increments the Gauge by 1. Use Add to increment it by arbitrary
	// values.
	Inc(...Label)
	// Dec decrements the Gauge by 1. Use Sub to decrement it by arbitrary
	// values.
	Dec(...Label)
	// Add adds the given value to the Gauge. (The value can be negative,
	// resulting in a decrease of the Gauge.)
	Add(float64, ...Label)
}

type GaugeOpts Opts

func NewGauge(opts GaugeOpts) Gauge {
	return &prometheusGauge{
		gauge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: opts.Subsystem,
			Name:      opts.Name,
			Help:      opts.Help,
		},
			opts.LabelNames,
		),
	}
}

type prometheusGauge struct {
	gauge *prometheus.GaugeVec
}

func (p prometheusGauge) Set(f float64, label ...Label) {
	p.gauge.With(toPrometheusLabel(label...)).Set(f)
}

func (p prometheusGauge) Inc(label ...Label) {
	p.gauge.With(toPrometheusLabel(label...)).Inc()
}

func (p prometheusGauge) Dec(label ...Label) {
	p.gauge.With(toPrometheusLabel(label...)).Dec()
}

func (p prometheusGauge) Add(f float64, label ...Label) {
	p.gauge.With(toPrometheusLabel(label...)).Add(f)
}
