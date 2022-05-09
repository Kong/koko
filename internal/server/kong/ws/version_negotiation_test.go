package ws

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kong/koko/internal/server/kong/ws/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Unittesting the individual negotiation function

func Test_rejectUnknownService(t *testing.T) {
	assert := assert.New(t)

	v, m := Negotiate("infundibulum", []string{"chrono-synclastic", "kitchen"})
	assert.Empty(v)
	assert.Equal("Unknown service.", m)

	v, m = Negotiate("cluster_protocol", []string{"avian", "tin-can"})
	assert.Empty(v)
	assert.Equal("no known version", m)
}

func Test_acceptVersion(t *testing.T) {
	assert := assert.New(t)

	// accept first option
	v, m := Negotiate("cluster_protocol", []string{"json"})
	assert.Equal("json", v)
	assert.NotEmpty(m)

	// or the second
	v, m = Negotiate("cluster_protocol", []string{"wrpc"})
	assert.Equal("wrpc", v)
	assert.NotEmpty(m)

	// prefer first option
	v, m = Negotiate("cluster_protocol", []string{"json", "wrpc"})
	assert.Equal("json", v)
	assert.NotEmpty(m)

	// even if it's not the first asked for
	v, m = Negotiate("cluster_protocol", []string{"wrpc", "json"})
	assert.Equal("json", v)
	assert.NotEmpty(m)

	// extra options are ignored
	v, m = Negotiate("cluster_protocol", []string{"x-ray", "json", "wrpc"})
	assert.Equal("json", v)
	assert.NotEmpty(m)
}

type dummyAuthenticator struct{}

func (dummyAuthenticator) Authenticate(r *http.Request) (*Manager, error) {
	vc, err := config.NewVersionCompatibilityProcessor(config.VersionCompatibilityOpts{
		Logger:        &zap.Logger{},
		KongCPVersion: config.KongGatewayCompatibilityVersion,
	})
	if err != nil {
		return nil, err
	}
	return NewManager(ManagerOpts{
		Cluster:                DefaultCluster{},
		DPVersionCompatibility: vc,
	}), nil
}

func dummyNegotiationHandler() NegotiationHandler {
	return NegotiationHandler{
		logger:        zap.NewNop(),
		authenticator: dummyAuthenticator{},
	}
}

func Test_rejectEmptyRequest(t *testing.T) {
	assert := assert.New(t)

	req := httptest.NewRequest("POST", "/version-handshake", strings.NewReader(""))
	rr := httptest.NewRecorder()

	dummyNegotiationHandler().ServeHTTP(rr, req)

	assert.Equal(400, rr.Code)
	assert.JSONEq(`{"message": "Invalid content type"}`, rr.Body.String())
}

func Test_rejectTextRequest(t *testing.T) {
	assert := assert.New(t)

	req := httptest.NewRequest("POST", "/version-handshake", strings.NewReader("Hello"))
	req.Header.Add("content-type", "plain/text")
	rr := httptest.NewRecorder()

	dummyNegotiationHandler().ServeHTTP(rr, req)

	assert.Equal(400, rr.Code)
	assert.JSONEq(`{"message": "Invalid content type"}`, rr.Body.String())
}

func Test_rejectEmptyJSONRequest(t *testing.T) {
	assert := assert.New(t)

	req := httptest.NewRequest("POST", "/version-handshake", strings.NewReader("{-}"))
	req.Header.Add("content-type", "application/json")
	rr := httptest.NewRecorder()

	dummyNegotiationHandler().ServeHTTP(rr, req)

	assert.Equal(400, rr.Code)

	req = httptest.NewRequest("POST", "/version-handshake", strings.NewReader("{}"))
	req.Header.Add("content-type", "application/json")
	rr = httptest.NewRecorder()

	dummyNegotiationHandler().ServeHTTP(rr, req)

	assert.Equal(400, rr.Code)
}

func Test_rejectIncompleteRequest(t *testing.T) {
	assert := assert.New(t)

	req := httptest.NewRequest("POST", "/version-handshake", strings.NewReader(`{
		"node": {
			"id": "me",
			"type": "KONG",
			"version": "1"
		}
	}`))
	req.Header.Add("content-type", "application/json")
	rr := httptest.NewRecorder()

	dummyNegotiationHandler().ServeHTTP(rr, req)

	assert.Equal(400, rr.Code)

	req = httptest.NewRequest("POST", "/version-handshake", strings.NewReader(`{
		"node": {
			"id": "me",
			"type": "KONG"
		},
		"services_requested": []
	}`))
	req.Header.Add("content-type", "application/json")
	rr = httptest.NewRecorder()

	dummyNegotiationHandler().ServeHTTP(rr, req)

	assert.Equal(400, rr.Code)

	req = httptest.NewRequest("POST", "/version-handshake", strings.NewReader(`{
		"node": {
			"id": "me",
			"type": "KONG",
			"version": "1"
		},
		"services_requested": [{"name": "cluster_protocol"}]
	}`))
	req.Header.Add("content-type", "application/json")
	rr = httptest.NewRecorder()

	dummyNegotiationHandler().ServeHTTP(rr, req)

	assert.Equal(400, rr.Code)

	req = httptest.NewRequest("POST", "/version-handshake", strings.NewReader(`{
		"node": {
			"id": "me",
			"type": "KONG",
			"version": "1"
		},
		"services_requested": [{"name": "cluster_protocol", "versions": []}]
	}`))
	req.Header.Add("content-type", "application/json")
	rr = httptest.NewRecorder()

	dummyNegotiationHandler().ServeHTTP(rr, req)

	assert.Equal(400, rr.Code)
}

type response struct {
	Node struct {
		ID string `json:"id"`
	} `json:"node"`
	ServicesAccepted []interface{} `json:"services_accepted"`
}

func Test_acceptFullRequest(t *testing.T) {
	assert := assert.New(t)

	req := httptest.NewRequest("POST", "/version-handshake", strings.NewReader(`{
		"node": {
			"id": "me",
			"type": "KONG",
			"version": "1"
		},
		"services_requested": [{"name": "cluster_protocol", "versions": ["json"]}]
	}`))
	req.Header.Add("content-type", "application/json")
	rr := httptest.NewRecorder()

	dummyNegotiationHandler().ServeHTTP(rr, req)

	assert.Equal(200, rr.Code)

	var resp response
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.Nil(err)
	assert.NotEmpty(resp.Node.ID)
	assert.Equal([]interface{}{
		map[string]interface{}{
			"name":    "cluster_protocol",
			"version": "json",
			"message": "Current",
		},
	}, resp.ServicesAccepted)
}
