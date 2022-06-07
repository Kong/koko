package ws

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/kong/go-wrpc/wrpc"
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
	BaseServices  Registerer
}

func NewHandler(opts HandlerOpts) (http.Handler, error) {
	mux := &http.ServeMux{}
	mux.Handle("/v1/outlet", handler{
		logger:        opts.Logger,
		authenticator: opts.Authenticator,
	})
	mux.Handle("/v1/wrpc", wrpcHandler{
		logger:        opts.Logger,
		authenticator: opts.Authenticator,
		baseServices:  opts.BaseServices,
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter,
		_ *http.Request,
	) {
		w.WriteHeader(http.StatusOK)
	})

	return mux, nil
}

type handler struct {
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

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (h handler) respondWithErr(w http.ResponseWriter, _ *http.Request,
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

type wrpcHandler struct {
	logger        *zap.Logger
	authenticator Authenticator
	baseServices  Registerer
}

func (h wrpcHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := validateRequest(r); err != nil {
		h.closeWithErr(w, http.StatusBadRequest, err)
		return
	}
	m, err := h.authenticator.Authenticate(r)
	if err != nil {
		h.closeWithErr(w, http.StatusBadRequest, err)
		return
	}
	peer := &wrpc.Peer{
		ErrLogger: func(err error) {
			h.logger.With(zap.Error(err), zap.String("wrpc-client-ip", r.RemoteAddr)).Error("peer object")
		},
	}
	err = h.baseServices.Register(peer)
	if err != nil {
		h.logger.With(zap.Error(err)).Error("register base wRPC services")
		return
	}
	err = peer.Upgrade(w, r)
	if err != nil {
		h.logger.With(zap.Error(err)).Error("upgrade to wRPC connection failed")
		return
	}

	queryParams := r.URL.Query()
	node := &Node{
		ID:       queryParams.Get(nodeIDKey),
		Hostname: queryParams.Get(nodeHostnameKey),
		Version:  queryParams.Get(nodeVersionKey),
		peer:     peer,
		logger:   h.logger.With(zap.String("wrpc-client-ip", r.RemoteAddr)),
	}

	// TODO: add the node somewhere until it's completed by some service
	_ = m
	_ = node
}

func (h wrpcHandler) closeWithErr(w http.ResponseWriter, statusCode int, errmsg error) {
	w.WriteHeader(statusCode)
	_, err := w.Write([]byte(errmsg.Error()))
	if err != nil {
		h.logger.With(zap.Error(err)).Error("Writing error to WebSocket")
	}
}
