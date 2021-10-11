package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{}

type HandlerOpts struct {
	Logger  *zap.Logger
	Manager *Manager
}

func NewHandler(opts HandlerOpts) (http.Handler, error) {
	mux := &http.ServeMux{}
	mux.Handle("/v1/outlet", Handler{
		logger:  opts.Logger,
		manager: opts.Manager,
	})
	return mux, nil
}

type Handler struct {
	logger  *zap.Logger
	manager *Manager
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.With(zap.Error(err)).Error("upgrade to websocket failed")
		return
	}
	h.manager.AddNode(Node{conn: c, logger: h.logger.With(zap.String(
		"client-ip", c.RemoteAddr().String()))})
}
