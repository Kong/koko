package health

import (
	"net/http"

	"github.com/kong/koko/internal/metrics"
)

type HandlerOpts struct{}

func NewHandler(_ HandlerOpts) (http.Handler, error) {
	return health{}, nil
}

type health struct{}

func (h health) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_ = metrics.Gauge("heartbeat", 1, metrics.Tag{Name: "server", Value: "health"})
	if r.URL.Path == "/health" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
