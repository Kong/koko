package ws

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
	if r.Method != "POST" {
		jsonErr(w, errMessage{"Invalid method"}, http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		jsonErr(w, errMessage{"Invalid content type"}, http.StatusBadRequest)
		return
	}

	m, err := h.authenticator.Authenticate(r)
	if err != nil {
		authErr, ok := err.(ErrAuth)
		if ok {
			jsonErr(w, errMessage{Message: authErr.Message}, authErr.HTTPStatus)
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
	resp.Node.ID = m.Cluster.Get()
	negVers := negotiatedVersions{}
	for _, serv := range resp.ServicesAccepted {
		negVers[serv.Name] = serv.Version
	}
	m.setNodeNegotiatedVersions(req.Node.ID, negVers)

	jsonBody, err := json.Marshal(resp)
	if err != nil {
		jsonErr(w, errMessage{Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	h.logger.Debug("encoded response", zap.Binary("jsonBody", jsonBody))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonBody)
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

	return inputSchema.Validate(req)
}

func jsonErr(w http.ResponseWriter, content interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	jsonBody, err := json.Marshal(content)
	if err != nil {
		jsonBody = []byte(`{"message": "Internal Error"}`)
		code = http.StatusInternalServerError
	}
	w.WriteHeader(code)
	_, _ = w.Write(jsonBody)
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
