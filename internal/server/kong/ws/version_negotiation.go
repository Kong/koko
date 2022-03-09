package ws

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"go.uber.org/zap"
)

var knownServices = map[string][]struct {
	version string
	message string
}{
	"cluster_protocol": {
		{version: "json", message: "Current"},
		{version: "wrpc", message: "beta"},
	},
}

type NegotiationHandler struct {
	logger        *zap.Logger
	authenticator Authenticator
}

type negotiatedVersions map[string]string

func (h NegotiationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" || r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, `{"message": "Invalid request"}`, http.StatusBadRequest)
		return
	}

	m, err := h.authenticator.Authenticate(r)
	if err != nil {
		autherr, ok := err.(ErrAuth)
		if ok {
			jsonErr(w, errMessage{Message: autherr.Message}, autherr.HTTPStatus)
		} else {
			jsonErr(w, errMessage{Message: "error while authenticating"}, http.StatusInternalServerError)
			h.logger.With(zap.Error(err)).Error("error while authenticating")
		}
		return
	}

	req, err := readJBody(*r)
	if err != nil {
		h.logger.With(zap.Error(err)).Error("bad request")
		jsonErr(w, errMessage{Message: "Invalid request: " + err.Error()}, http.StatusBadRequest)
		return
	}

	h.logger.Debug("decoded request", zap.Any("req", req))

	resp := negotiateServices(req)
	negVers := negotiatedVersions{}
	for _, serv := range resp.ServicesAccepted {
		negVers[serv.Name] = serv.Version
	}
	m.setNodeNegotiatedVersions(req.Node.ID, negVers)

	jsonbody, err := json.Marshal(resp)
	if err != nil {
		jsonErr(w, errMessage{Message: err.Error()}, http.StatusBadRequest)
		return
	}
	h.logger.Debug("encoded response", zap.Binary("jsonbody", jsonbody))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonbody)
	if err != nil {
		h.logger.With(zap.Error(err)).Error("error writing response")
	}
}

type versionRequest struct {
	Node struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Version  string `json:"version"`
		Hostname string `json:"hostname"`
	} `json:"node"`

	ServicesRequested []struct {
		Name     string   `json:"name"`
		Versions []string `json:"versions"`
	} `json:"services_requested"`

	Metadata interface{} `json:"metadata"`
}

func readJBody(r http.Request) (versionRequest, error) {
	if r.Method != "POST" || r.Header.Get("Content-Type") != "application/json" {
		return versionRequest{}, fmt.Errorf("invalid content")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return versionRequest{}, err
	}

	var j interface{}
	err = json.Unmarshal(body, &j)
	if err != nil {
		return versionRequest{}, err
	}

	err = validateVersionRequest(j)
	if err != nil {
		return versionRequest{}, err
	}

	var v versionRequest
	err = json.Unmarshal(body, &v)
	if err != nil {
		return versionRequest{}, err
	}

	return v, nil
}

var inputSchema *jsonschema.Schema

const (
	inputSchemaSource = `{
		"type": "object",
		"properties": {
			"node": {
				"type": "object",
				"properties": {
					"id": { "type": "string" },
					"type": { "type": "string", "pattern": "^KONG$" },
					"version": { "type": "string" },
					"hostname": { "type": "string" }
				},
				"required": ["id", "type", "version"]
			},
			"services_requested": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"name": { "type": "string" },
						"versions": {
							"type": "array",
							"items": { "type": "string" },
							"minItems": 1
						}
					},
					"required": ["name", "versions"]
				}
			},
			"metadata": { "type": "object" }
		},
		"required": ["node", "services_requested"]
	}`
	inputSchemaBase = `/kong-version-negotiation`
)

func validateVersionRequest(req interface{}) error {
	if inputSchema == nil {
		inputSchema = jsonschema.MustCompileString(inputSchemaBase, inputSchemaSource)
	}

	err := inputSchema.Validate(req)
	if err != nil {
		return err
	}

	return nil
}

func jsonErr(w http.ResponseWriter, content interface{}, code int) {
	jsonbody, err := json.Marshal(content)
	if err != nil {
		http.Error(w, "Very bad content", code)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write(jsonbody)
}

type errMessage struct {
	Message string `json:"message"`
}

type versionNode struct {
	ID string `json:"id"`
}

type versionServiceAccepted struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Message string `json:"message,omitempty"`
}

type versionServiceRejected struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

type versionResponse struct {
	Node             versionNode              `json:"node"`
	ServicesAccepted []versionServiceAccepted `json:"services_accepted"`
	ServicesRejected []versionServiceRejected `json:"services_rejected"`
}

func negotiateServices(req versionRequest) versionResponse {
	resp := versionResponse{
		Node:             versionNode{ID: uuid.New().String()},
		ServicesAccepted: []versionServiceAccepted{},
		ServicesRejected: []versionServiceRejected{},
	}

	for _, requestedService := range req.ServicesRequested {
		vers, msg := Negotiate(requestedService.Name, requestedService.Versions)
		if vers == "" {
			resp.ServicesRejected = append(resp.ServicesRejected, versionServiceRejected{
				Name:    requestedService.Name,
				Message: msg,
			})
		} else {
			resp.ServicesAccepted = append(resp.ServicesAccepted, versionServiceAccepted{
				Name:    requestedService.Name,
				Version: vers,
				Message: msg,
			})
		}
	}

	return resp
}

func Negotiate(service string, versions []string) (vers string, msg string) {
	knownVersions, ok := knownServices[service]
	if !ok {
		return "", "Unknown service."
	}

	for _, knownVers := range knownVersions {
		for _, inputVers := range versions {
			if inputVers == knownVers.version {
				return knownVers.version, knownVers.message
			}
		}
	}

	return "", "no known version"
}
