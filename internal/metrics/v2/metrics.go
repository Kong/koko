package v2

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Label struct {
	Key   string
	Value string
}

func PrometheusHandler() http.Handler {
	m := http.NewServeMux()
	m.Handle("/metrics", promhttp.Handler())
	return m
}
