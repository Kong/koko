package health

import (
	"net/http"

	"github.com/segmentio/stats/v4"
)

type HandlerOpts struct{}

func NewHandler(_ HandlerOpts) (http.Handler, error) {
	return health{}, nil
}

type health struct{}

func (h health) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stats.Set("heartbeat", 1, stats.Tag{Name: "server", Value: "health"})
	if r.URL.Path == "/health" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
