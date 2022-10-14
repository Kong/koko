package v2

import "github.com/prometheus/client_golang/prometheus"

type Opts struct {
	Subsystem string
	Name      string

	// Help provides information about this metric.
	//
	// Metrics with the same fully-qualified name must have the same Help
	// string.
	Help string

	LabelNames []string
}

const namespace = "kong"

func toPrometheusLabel(label ...Label) prometheus.Labels {
	labels := prometheus.Labels{}
	for _, l := range label {
		labels[l.Key] = l.Value
	}
	return labels
}
