package v2

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Histogram is a metrics histogram.
// Please consult Prometheus documentation for a detailed understanding.
type Histogram interface {
	// Observe adds a single observation to the histogram. Observations are
	// usually positive or zero. Negative observations are accepted but
	// prevent current versions of Prometheus from properly detecting
	// counter resets in the sum of observations. See
	// https://prometheus.io/docs/practices/histograms/#count-and-sum-of-observations
	// for details.
	Observe(float64, ...Label)
}

type HistogramOpts struct {
	Subsystem string
	Name      string

	// Help provides information about this Histogram.
	//
	// Metrics with the same fully-qualified name must have the same Help
	// string.
	Help string

	// Buckets defines the buckets into which observations are counted. Each
	// element in the slice is the upper inclusive bound of a bucket. The
	// values must be sorted in strictly increasing order. There is no need
	// to add a highest bucket with +Inf bound, it will be added
	// implicitly. The default value is DefBuckets.
	Buckets []float64

	LabelNames []string
}

func NewHistogram(opts HistogramOpts) Histogram {
	return &prometheusHistogram{
		histogram: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: opts.Subsystem,
			Name:      opts.Name,
			Help:      opts.Help,
		},
			opts.LabelNames,
		),
	}
}

type prometheusHistogram struct {
	histogram *prometheus.HistogramVec
}

func (p prometheusHistogram) Observe(f float64, label ...Label) {
	p.histogram.With(toPrometheusLabel(label...)).Observe(f)
}
