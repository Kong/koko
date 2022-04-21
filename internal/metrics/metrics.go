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
	Datadog
	Prometheus
)

const (
	metricPrefix = "koko."
)

var validClients = map[string]Client{
	"noop":       NoOpClient,
	"datadog":    Datadog,
	"prometheus": Prometheus,
}

func ParseClient(client string) (Client, error) {
	if client == "" {
		return NoOpClient, nil
	}

	if c, ok := validClients[client]; ok {
		return c, nil
	}
	return NoOpClient, fmt.Errorf("invalid metrics_client '%s'", client)
}

func (c Client) String() string {
	for k, client := range validClients {
		if client == c {
			return k
		}
	}
	panic("invalid client")
}

func InitStatsClient(logger *zap.Logger, client Client) http.Handler {
	switch client {
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
