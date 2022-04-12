package config

import (
	"fmt"
)

type MetricsClient int

const (
	NoOpClient MetricsClient = iota
	StatsD
	Datadog
	Prometheus
)

var validMetricClients = []string{"noop", "statsd", "datadog", "prometheus"}

func ParseMetricsClient(client string) (MetricsClient, error) {
	switch client {
	case validMetricClients[0], "":
		return NoOpClient, nil
	case validMetricClients[1]:
		return StatsD, nil
	case validMetricClients[2]:
		return Datadog, nil
	case validMetricClients[3]:
		return Prometheus, nil
	default:
		return NoOpClient, fmt.Errorf("invalid metrics_client '%s'", client)
	}
}

func (c MetricsClient) String() string {
	return validMetricClients[c]
}
