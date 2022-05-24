package ws

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/kong/koko/internal/json"
	"go.uber.org/zap"
)

const (
	nodeIDKey       = "node_id"
	nodeHostnameKey = "node_hostname"
	nodeVersionKey  = "node_version"
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
		_ *http.Request,
	) {
		w.WriteHeader(http.StatusOK)
	})

	return mux, nil
}

type Handler struct {
	logger        *zap.Logger
	authenticator Authenticator
}

func validateRequest(r *http.Request) error {
	queryParams := r.URL.Query()
	if queryParams.Get(nodeIDKey) == "" {
		return fmt.Errorf("invalid request, missing node_id query parameter")
	}
	if queryParams.Get(nodeHostnameKey) == "" {
		return fmt.Errorf("invalid request, " +
			"missing node_hostname query parameter")
	}
	if queryParams.Get(nodeVersionKey) == "" {
		return fmt.Errorf("invalid request, " +
			"missing node_version query parameter")
	}
	return nil
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := validateRequest(r); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			h.logger.With(zap.Error(err)).Error(
				"write bad request response for websocket upgrade")
		}
		h.logger.Info("received invalid websocket upgrade request from DP",
			zap.Error(err))
		return
	}

	queryParams := r.URL.Query()
	nodeID := queryParams.Get(nodeIDKey)
	nodeHostname := queryParams.Get(nodeHostnameKey)
	nodeVersion := queryParams.Get(nodeVersionKey)
	loggerWithNode := h.logger.With(
		zap.String("node-id", nodeID),
		zap.String("node-hostname", nodeHostname),
		zap.String("node-version", nodeVersion),
	)

	m, err := h.authenticator.Authenticate(r)
	if err != nil {
		h.respondWithErr(w, r, err)
		loggerWithNode.Error("failed to authenticate DP node", zap.Error(err))
		return
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		loggerWithNode.Error("failed to upgrade websocket connection", zap.Error(err))
		return
	}

	node := &Node{
		ID:       nodeID,
		Hostname: nodeHostname,
		Version:  nodeVersion,
		conn:     c,
	}
	node.logger = nodeLogger(node, m.logger)
	m.AddNode(node)
}

func (h Handler) respondWithErr(w http.ResponseWriter, _ *http.Request,
	err error,
) {
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
