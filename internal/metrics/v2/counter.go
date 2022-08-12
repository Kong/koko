package v2

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Counter is a metrics gauge.
// Please consult Prometheus documentation for a detailed understanding.
type Counter interface {
	// Inc increments the gauge by 1. Use Add to increment it by arbitrary
	// non-negative values.
	Inc(...Label)
	// Add adds the given value to the gauge. It panics if the value is <
	// 0.
	Add(float64, ...Label)
}

type CounterOpts Opts

func NewCounter(opts CounterOpts) Counter {
	return &prometheusCounter{
		counter: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: opts.Subsystem,
			Name:      opts.Name,
			Help:      opts.Help,
		},
			opts.LabelNames,
		),
	}
}

type prometheusCounter struct {
	counter *prometheus.CounterVec
}

func (p prometheusCounter) Inc(label ...Label) {
	p.counter.With(toPrometheusLabel(label...)).Inc()
}

func (p prometheusCounter) Add(f float64, label ...Label) {
	p.counter.With(toPrometheusLabel(label...)).Add(f)
}
