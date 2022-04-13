package metrics

import (
	"fmt"
	"net/http"
	"os"

	"github.com/segmentio/stats/v4"
	"github.com/segmentio/stats/v4/datadog"
	"github.com/segmentio/stats/v4/prometheus"
	"go.uber.org/zap"
)

type Client int

const (
	NoOpClient Client = iota
	StatsD
	Datadog
	Prometheus
)

const (
	metricPrefix = "koko."
)

var validClients = []string{"noop", "statsd", "datadog", "prometheus"}

func ParseClient(client string) (Client, error) {
	switch client {
	case validClients[0], "":
		return NoOpClient, nil
	case validClients[1]:
		return StatsD, nil
	case validClients[2]:
		return Datadog, nil
	case validClients[3]:
		return Prometheus, nil
	default:
		return NoOpClient, fmt.Errorf("invalid metrics_client '%s'", client)
	}
}

func (c Client) String() string {
	return validClients[c]
}

func InitStatsClient(logger *zap.Logger, client Client) http.Handler {
	switch client {
	case StatsD:
		host := os.Getenv("STATSD_HOST")
		if host == "" {
			panic("StatsD environment variable 'STATSD_HOST' must be set")
		}
		stats.DefaultEngine = stats.NewEngine(metricPrefix, datadog.NewClient(host))
		fallthrough
	case Datadog:
		ddEnvTagsMapping := []struct{ envVar, tagName string }{
			{"DD_ENTITY_ID", "dd.internal.entity_id"},
			{"DD_ENV", "env"},
			{"DD_SERVICE", "service"},
			{"DD_VERSION", "version"},
		}
		tags := make([]stats.Tag, 0, len(ddEnvTagsMapping))
		for _, ddenv := range ddEnvTagsMapping {
			if v := os.Getenv(ddenv.envVar); v != "" {
				tags = append(tags, stats.Tag{Name: ddenv.tagName, Value: v})
			}
		}

		agent := os.Getenv("DD_AGENT_HOST")
		if agent == "" {
			panic("Datadog client environment variable 'DD_AGENT_HOST' must be set")
		}

		stats.DefaultEngine = stats.NewEngine(metricPrefix, datadog.NewClient(agent), tags...)
	case Prometheus:
		stats.DefaultEngine = stats.NewEngine(metricPrefix, prometheus.DefaultHandler)
		return prometheus.DefaultHandler
	case NoOpClient:
		fallthrough
	default:
		logger.Info("metrics client config not set")
	}
	return nil
}
