package health

import (
	"net/http"
)

type HandlerOpts struct{}

func NewHandler(_ HandlerOpts) (http.Handler, error) {
	return health{}, nil
}

type health struct{}

func (h health) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/health" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
