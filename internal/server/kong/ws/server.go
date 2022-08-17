package ws

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/gorilla/websocket"
	"github.com/kong/go-wrpc/wrpc"
	"github.com/kong/koko/internal/json"
	"go.uber.org/zap"
)

const (
	nodeIDKey       = "node_id"
	nodeHostnameKey = "node_hostname"
	nodeVersionKey  = "node_version"

	minimalAcceptableDPVersion = "2.5.0"
)

var upgrader = websocket.Upgrader{}

type HandlerOpts struct {
	Logger        *zap.Logger
	Authenticator Authenticator
	BaseServices  Registerer
}

func NewHandler(opts HandlerOpts) (http.Handler, error) {
	mux := &http.ServeMux{}
	mux.Handle("/v1/outlet", websocketHandler{
		logger:        opts.Logger,
		authenticator: opts.Authenticator,
	})

	if opts.BaseServices != nil {
		mux.Handle("/v1/wrpc", wrpcHandler{
			websocketHandler: websocketHandler{
				logger:        opts.Logger,
				authenticator: opts.Authenticator,
			},
			baseServices: opts.BaseServices,
		})
	}

	mux.HandleFunc("/health", func(w http.ResponseWriter,
		_ *http.Request,
	) {
		w.WriteHeader(http.StatusOK)
	})

	return mux, nil
}

type websocketHandler struct {
	logger        *zap.Logger
	authenticator Authenticator
}

// getFrontNumber returns the first from a string of numbers
// separated by dots.
// The rest of the string is returned too for further processing.
// Any non-number is considered zero and ends the sequence.
func getFrontNumber(s string) (uint64, string) {
	const (
		decimalBase   = 10
		uint64BitSize = 64
	)

	head, tail, _ := strings.Cut(s, ".")
	n, err := strconv.ParseUint(head, decimalBase, uint64BitSize)
	if err != nil {
		return 0, ""
	}
	return n, tail
}

// trivialVersionParse tries to parse up to the three first segments
// of a version string.
func trivialVersionParse(s string) semver.Version {
	var version semver.Version
	version.Major, s = getFrontNumber(s)
	version.Minor, s = getFrontNumber(s)
	version.Patch, _ = getFrontNumber(s)
	return version
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
	nodeVersion := queryParams.Get(nodeVersionKey)
	if nodeVersion == "" {
		return fmt.Errorf("invalid request, " +
			"missing node_version query parameter")
	}
	if trivialVersionParse(nodeVersion).LT(trivialVersionParse(minimalAcceptableDPVersion)) {
		return fmt.Errorf("unsupported dataplane version")
	}

	return nil
}

func (h websocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	node, err := NewNode(nodeOpts{
		id:         nodeID,
		hostname:   nodeHostname,
		version:    nodeVersion,
		connection: c,
		logger:     m.logger.With(zap.String("client-ip", r.RemoteAddr)),
	})
	if err != nil {
		h.logger.Error("Create websocket Node failed", zap.Error(err), zap.String("client-ip", r.RemoteAddr))
		h.respondWithErr(w, r, err)
		return
	}
	m.AddWebsocketNode(node)
}

// respondWithErr sends an error HTTP response, with a json message.
func (h websocketHandler) respondWithErr(w http.ResponseWriter, _ *http.Request,
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
	websocketHandler
	baseServices Registerer
}

func (h wrpcHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := validateRequest(r); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			h.logger.Error("write bad request response for websocket upgrade", zap.Error(err))
		}
		h.logger.Info("received invalid websocket upgrade request from DP", zap.Error(err))
		return
	}
	m, err := h.authenticator.Authenticate(r)
	if err != nil {
		h.respondWithErr(w, r, err)
		return
	}
	queryParams := r.URL.Query()

	var node *Node
	peer := &wrpc.Peer{
		ErrLogger: func(err error) {
			node.Logger.Error("wRPC Peer object error", zap.Error(err))
		},
		ClosedCallbackFunc: func(p *wrpc.Peer) {
			if node != nil {
				m.removeNode(node)
			}
		},
	}

	node, err = NewNode(nodeOpts{
		id:       queryParams.Get(nodeIDKey),
		hostname: queryParams.Get(nodeHostnameKey),
		version:  queryParams.Get(nodeVersionKey),
		logger:   m.logger.With(zap.String("wrpc-client-ip", r.RemoteAddr)),
		peer:     peer,
	})
	if err != nil {
		h.logger.Error("Create wRPC Node failed", zap.Error(err), zap.String("wrpc-client-ip", r.RemoteAddr))
		h.respondWithErr(w, r, err)
		return
	}

	err = h.baseServices.Register(node.peer)
	if err != nil {
		node.Logger.Error("register base wRPC services", zap.Error(err))
		return
	}
	err = node.peer.Upgrade(w, r)
	if err != nil {
		node.Logger.Error("upgrade to wRPC connection failed", zap.Error(err))
		h.respondWithErr(w, r, err)
		return
	}

	if err = m.AddPendingNode(node); err != nil {
		node.Logger.Error("adding to pending node list", zap.Error(err))
		if err = node.Close(); err != nil {
			node.Logger.Error("closing node", zap.Error(err))
		}
	}
}
