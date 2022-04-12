package config

type MetricsClient int

const (
	NoOpClient MetricsClient = iota
	StatsD
	Datadog
	Prometheus
)

func (c MetricsClient) String() string {
	return [...]string{"noop", "statsd", "datadog", "prometheus"}[c]
}
