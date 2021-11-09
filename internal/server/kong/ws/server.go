package ws

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{}

type HandlerOpts struct {
	Logger        *zap.Logger
	Authenticator Authenticator
}

func NewHandler(opts HandlerOpts) (http.Handler, error) {
	mux := &http.ServeMux{}
	mux.Handle("/v1/outlet", Handler{
		logger:        opts.Logger,
		authenticator: opts.Authenticator,
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter,
		_ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return mux, nil
}

type Handler struct {
	logger        *zap.Logger
	authenticator Authenticator
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m, err := h.authenticator.Authenticate(r)
	if err != nil {
		h.respondWithErr(w, r, err)
		return
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.With(zap.Error(err)).Error("upgrade to websocket failed")
		return
	}
	m.AddNode(Node{conn: c, logger: h.logger.With(zap.String(
		"client-ip", c.RemoteAddr().String()))})
}

func (h Handler) respondWithErr(w http.ResponseWriter, _ *http.Request,
	err error) {
	authErr, ok := err.(ErrAuth)
	if ok {
		resp, err := json.Marshal(map[string]string{"message": authErr.Message})
		if err != nil {
			h.logger.With(zap.Error(err)).Error("marshal JSON")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(authErr.HTTPStatus)
		_, err = w.Write(resp)
		if err != nil {
			h.logger.With(zap.Error(err)).Error("write auth error")
		}
		return
	}
	h.logger.With(zap.Error(err)).Error("error while authenticating")
	w.WriteHeader(http.StatusInternalServerError)
}
