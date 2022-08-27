package admin

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func goodService() *v1.Service {
	return &v1.Service{
		Name: "foo",
		Host: "example.com",
		Path: "/",
	}
}

func validateGoodService(body *httpexpect.Object) {
	body.ContainsKey("id")
	body.NotContainsKey("url")
	body.ContainsKey("created_at")
	body.ContainsKey("updated_at")
	body.ValueEqual("write_timeout", 60000)
	body.ValueEqual("read_timeout", 60000)
	body.ValueEqual("connect_timeout", 60000)
	body.ValueEqual("name", "foo")
	body.ValueEqual("path", "/")
	body.ValueEqual("host", "example.com")
	body.ValueEqual("port", 80)
	body.ValueEqual("retries", 5)
	body.ValueEqual("protocol", "http")
	body.ValueEqual("enabled", true)
}

func TestServiceCreate(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	t.Run("creates a valid service", func(t *testing.T) {
		res := c.POST("/v1/services").WithJSON(goodService()).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		validateGoodService(body)
	})
	t.Run("creating a service with protocol=ws fails", func(t *testing.T) {
		util.SkipTestIfEnterpriseTesting(t, true)
		service := goodService()
		service.Protocol = typedefs.ProtocolWS
		res := c.POST("/v1/services").WithJSON(service).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		resErr := body.Value("details").Array().Element(0)
		resErr.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.String())
		resErr.Object().ValueEqual("messages", []string{
			"'ws' and 'wss' protocols are Kong Enterprise-only features. " +
				"Please upgrade to Kong Enterprise to use this feature.",
		})
	})
	t.Run("creating a service with protocol=wss fails", func(t *testing.T) {
		util.SkipTestIfEnterpriseTesting(t, true)
		service := goodService()
		service.Protocol = typedefs.ProtocolWSS
		res := c.POST("/v1/services").WithJSON(service).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		resErr := body.Value("details").Array().Element(0)
		resErr.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.String())
		resErr.Object().ValueEqual("messages", []string{
			"'ws' and 'wss' protocols are Kong Enterprise-only features. " +
				"Please upgrade to Kong Enterprise to use this feature.",
		})
	})
	t.Run("creates a valid service with enabled=false", func(t *testing.T) {
		service := goodService()
		service.Name = "disabled-svc"
		service.Enabled = wrapperspb.Bool(false)
		serviceJSON, err := json.ProtoJSONMarshal(service)
		require.Nil(t, err)
		res := c.POST("/v1/services").WithBytes(serviceJSON).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.Value("enabled").Equal(false)
	})
	t.Run("creates a valid service with url set", func(t *testing.T) {
		service := &v1.Service{
			Name: "with-url-svc",
			Url:  "https://example.com:8080/sample/path",
		}
		serviceJSON, err := json.ProtoJSONMarshal(service)
		require.Nil(t, err)
		res := c.POST("/v1/services").WithBytes(serviceJSON).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.NotContainsKey("url")
		body.Value("name").Equal(service.Name)
		body.Value("protocol").Equal("https")
		body.Value("host").Equal("example.com")
		body.Value("port").Equal(8080)
		body.Value("path").Equal("/sample/path")
	})
	t.Run("creates a valid service with url without port set", func(t *testing.T) {
		service := &v1.Service{
			Name: "with-url-without-port-svc",
			Url:  "https://foo/bar",
		}
		serviceJSON, err := json.ProtoJSONMarshal(service)
		require.Nil(t, err)
		res := c.POST("/v1/services").WithBytes(serviceJSON).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.NotContainsKey("url")
		body.Value("name").Equal(service.Name)
		body.Value("protocol").Equal("https")
		body.Value("host").Equal("foo")
		body.Value("port").Equal(443)
		body.Value("path").Equal("/bar")
	})
	t.Run("creates a service with invalid url set fails", func(t *testing.T) {
		service := &v1.Service{
			Name: "invalid",
			Url:  "foo.com",
		}
		serviceJSON, err := json.ProtoJSONMarshal(service)
		require.Nil(t, err)
		res := c.POST("/v1/services").WithBytes(serviceJSON).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		gotErr := body.Value("details").Array().Element(0)
		gotErr.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.
			String())
		gotErr.Object().ValueEqual("messages", []string{
			"missing properties: 'host'",
		})
	})
	t.Run("creates a service with invalid protocol fails", func(t *testing.T) {
		service := &v1.Service{
			Name: "invalid",
			Url:  "ftp://foo.com",
		}
		serviceJSON, err := json.ProtoJSONMarshal(service)
		require.Nil(t, err)
		res := c.POST("/v1/services").WithBytes(serviceJSON).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		gotErr := body.Value("details").Array().Element(0)
		gotErr.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_FIELD.
			String())
		gotErr.Object().ValueEqual("messages", []string{
			`value must be one of "http", "https", "grpc", ` +
				`"grpcs", "tcp", "udp", "tls", "tls_passthrough", "ws", "wss"`,
		})
	})
	t.Run("creates a service with url set but missing path successfully", func(t *testing.T) {
		service := &v1.Service{
			Name: "invalid",
			Url:  "https://foo",
		}
		serviceJSON, err := json.ProtoJSONMarshal(service)
		require.Nil(t, err)
		res := c.POST("/v1/services").WithBytes(serviceJSON).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.NotContainsKey("url")
		body.Value("name").Equal(service.Name)
		body.Value("protocol").Equal("https")
		body.Value("host").Equal("foo")
		body.Value("port").Equal(443)
		body.NotContainsKey("path")
	})
	t.Run("creating invalid service fails with 400", func(t *testing.T) {
		svc := &v1.Service{
			Name:           "foo",
			Host:           "example.com",
			Path:           "//foo", // invalid '//' sequence
			ConnectTimeout: -2,      // invalid timeout
		}
		res := c.POST("/v1/services").WithJSON(svc).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(2)
		errs := body.Value("details").Array()
		var fields []string
		for _, err := range errs.Iter() {
			err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_FIELD.
				String())
			fields = append(fields, err.Object().Value("field").String().Raw())
		}
		require.ElementsMatch(t, []string{"path", "connect_timeout"}, fields)
	})
	t.Run("recreating the service with the same name fails", func(t *testing.T) {
		svc := goodService()
		res := c.POST("/v1/services").WithJSON(svc).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "data constraint error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0)
		err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.
			String())
		err.Object().ValueEqual("field", "name")
	})
	t.Run("recreating the service with the same id fails", func(t *testing.T) {
		sid := uuid.NewString()
		svc := goodService()
		svc.Name = "same-id-service"
		svc.Id = sid
		res := c.POST("/v1/services").WithJSON(svc).Expect()
		res.Status(http.StatusCreated)

		res = c.POST("/v1/services").WithJSON(svc).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "data constraint error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0)
		err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.
			String())
		err.Object().ValueEqual("field", "id")
	})
	t.Run("service with a '-' in name can be created", func(t *testing.T) {
		svc := goodService()
		svc.Name = "foo-with-dash"
		res := c.POST("/v1/services").WithJSON(svc).Expect()
		res.Status(http.StatusCreated)
	})
	t.Run("creates a valid service specifying the ID using POST", func(t *testing.T) {
		service := goodService()
		service.Name = "with-id"
		service.Id = uuid.NewString()
		res := c.POST("/v1/services").WithJSON(service).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.Value("id").Equal(service.Id)
		body.Value("name").Equal(service.Name)
	})
	t.Run("creates a service referencing a not-existing CA cert fails", func(t *testing.T) {
		service := goodService()
		service.Protocol = "https"
		service.Name = "with-cert"
		service.CaCertificates = []string{uuid.NewString()}
		res := c.POST("/v1/services").WithJSON(service).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "data constraint error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0)
		err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.
			String())
		err.Object().ValueEqual("field", "ca_certificates")
	})
	t.Run("creates a valid service referencing a CA cert successfully", func(t *testing.T) {
		// create CA certificate
		res := c.POST("/v1/ca-certificates").WithJSON(&v1.CACertificate{
			Cert: goodCACertOne,
		}).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.Value("cert").String().Equal(goodCACertOne)
		caCertID := body.Value("id").String().Raw()

		// create service
		service := goodService()
		service.Name = "with-good-cert"
		service.Protocol = "https"
		service.CaCertificates = []string{caCertID}
		res = c.POST("/v1/services").WithJSON(service).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body = res.JSON().Path("$.item").Object()
		body.Value("ca_certificates").Array().Length().Equal(1)
		gotCACertID := body.Value("ca_certificates").Array().Element(0).String().Raw()
		require.True(t, caCertID == gotCACertID)
	})
	t.Run("creates a valid service referencing a client cert successfully", func(t *testing.T) {
		// create certificate
		res := c.POST("/v1/certificates").WithJSON(&v1.Certificate{
			Cert: goodCertOne,
			Key:  goodKeyOne,
		}).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.Value("cert").String().Equal(goodCertOne)
		body.Value("key").String().Equal(goodKeyOne)
		certID := body.Value("id").String().Raw()

		// create service
		service := goodService()
		service.Name = "with-good-client-cert"
		service.Protocol = "https"
		service.ClientCertificate = &v1.Certificate{
			Id: certID,
		}
		res = c.POST("/v1/services").WithJSON(service).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body = res.JSON().Path("$.item").Object()
		gotCertID := body.Value("client_certificate").Object().Path("$.id").String().Raw()
		require.True(t, certID == gotCertID)
	})
	t.Run("ensure failure creating a service referencing a non-existing client cert",
		func(t *testing.T) {
			service := goodService()
			service.Name = "with-bad-client-cert"
			service.Protocol = "https"
			service.ClientCertificate = &v1.Certificate{
				Id: uuid.NewString(),
			}
			res := c.POST("/v1/services").WithJSON(service).Expect()
			res.Status(http.StatusBadRequest)
			body := res.JSON().Object()
			body.ValueEqual("message", "data constraint error")
			body.Value("details").Array().Length().Equal(1)
			err := body.Value("details").Array().Element(0)
			err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.
				String())
			err.Object().ValueEqual("field", "client_certificate.id")
		})
	t.Run("ensure failure creating a service referencing a client cert without https",
		func(t *testing.T) {
			util.SkipTestIfEnterpriseTesting(t, true)
			service := goodService()
			service.Name = "with-bad-protocol-client-cert"
			service.Protocol = "http"
			service.ClientCertificate = &v1.Certificate{
				Id: uuid.NewString(),
			}
			res := c.POST("/v1/services").WithJSON(service).Expect()
			res.Status(http.StatusBadRequest)
			body := res.JSON().Object()
			body.ValueEqual("message", "validation error")
			body.Value("details").Array().Length().Equal(1)
			err := body.Value("details").Array().Element(0)
			err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.
				String())
			err.Object().ValueEqual("messages", []string{
				"client_certificate can be set only when protocol is `https`",
			})
		})
	t.Run("ensure deleting a service referencing a client cert doesn't remove the cert itself",
		func(t *testing.T) {
			// create certificate
			res := c.POST("/v1/certificates").WithJSON(&v1.Certificate{
				Cert: goodCertTwo,
				Key:  goodKeyTwo,
			}).Expect()
			res.Status(http.StatusCreated)
			res.Header("grpc-metadata-koko-status-code").Empty()
			body := res.JSON().Path("$.item").Object()
			body.Value("cert").String().Equal(goodCertTwo)
			body.Value("key").String().Equal(goodKeyTwo)
			certID := body.Value("id").String().Raw()

			// create service
			service := goodService()
			service.Name = "delete-svc-with-client-cert"
			service.Protocol = "https"
			service.ClientCertificate = &v1.Certificate{
				Id: certID,
			}
			res = c.POST("/v1/services").WithJSON(service).Expect()
			res.Status(http.StatusCreated)
			res.Header("grpc-metadata-koko-status-code").Empty()
			body = res.JSON().Path("$.item").Object()
			gotCertID := body.Value("client_certificate").Object().Path("$.id").String().Raw()
			require.True(t, certID == gotCertID)
			svcID := body.Value("id").String().Raw()

			// delete service
			c.DELETE("/v1/services/" + svcID).Expect().Status(http.StatusNoContent)

			// check certificate still exists
			c.GET("/v1/certificates/" + certID).Expect().Status(http.StatusOK)
		})
	t.Run("ensure deleting a cert referenced in a service removes the services itself too",
		func(t *testing.T) {
			// create certificate
			res := c.POST("/v1/certificates").WithJSON(&v1.Certificate{
				Cert: goodCertThree,
				Key:  goodKeyThree,
			}).Expect()
			res.Status(http.StatusCreated)
			res.Header("grpc-metadata-koko-status-code").Empty()
			body := res.JSON().Path("$.item").Object()
			body.Value("cert").String().Equal(goodCertThree)
			body.Value("key").String().Equal(goodKeyThree)
			certID := body.Value("id").String().Raw()

			// create service
			service := goodService()
			service.Name = "cascade-delete-with-client-cert"
			service.Protocol = "https"
			service.ClientCertificate = &v1.Certificate{
				Id: certID,
			}
			res = c.POST("/v1/services").WithJSON(service).Expect()
			res.Status(http.StatusCreated)
			res.Header("grpc-metadata-koko-status-code").Empty()
			body = res.JSON().Path("$.item").Object()
			gotCertID := body.Value("client_certificate").Object().Path("$.id").String().Raw()
			require.True(t, certID == gotCertID)
			svcID := body.Value("id").String().Raw()

			// delete certificate
			c.DELETE("/v1/certificates/" + certID).Expect().Status(http.StatusNoContent)

			// check certificate doesn't exist anymore
			c.GET("/v1/services/" + svcID).Expect().Status(http.StatusNotFound)
		})
}

func TestServiceUpsert(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	t.Run("upserts a valid service", func(t *testing.T) {
		res := c.PUT("/v1/services/" + uuid.NewString()).
			WithJSON(goodService()).
			Expect()
		res.Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		validateGoodService(body)
	})
	t.Run("upserting an invalid service fails with 400", func(t *testing.T) {
		svc := &v1.Service{
			Name:           "foo",
			Host:           "example.com",
			Path:           "//foo", // invalid '//' sequence
			ConnectTimeout: -2,      // invalid timeout
		}
		res := c.PUT("/v1/services/" + uuid.NewString()).
			WithJSON(svc).
			Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(2)
		errs := body.Value("details").Array()
		var fields []string
		for _, err := range errs.Iter() {
			err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_FIELD.
				String())
			fields = append(fields, err.Object().Value("field").String().Raw())
		}
		require.ElementsMatch(t, []string{"path", "connect_timeout"}, fields)
	})
	t.Run("recreating the service with the same name but different id fails",
		func(t *testing.T) {
			svc := goodService()
			res := c.PUT("/v1/services/" + uuid.NewString()).
				WithJSON(svc).
				Expect()
			res.Status(http.StatusBadRequest)
			body := res.JSON().Object()
			body.ValueEqual("message", "data constraint error")
			body.Value("details").Array().Length().Equal(1)
			err := body.Value("details").Array().Element(0)
			err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.
				String())
			err.Object().ValueEqual("field", "name")
		})
	t.Run("upsert correctly updates a route", func(t *testing.T) {
		sid := uuid.NewString()
		svc := &v1.Service{
			Id:   sid,
			Name: "r1",
			Host: "example.com",
			Path: "/bar",
		}
		res := c.POST("/v1/services").
			WithJSON(svc).
			Expect()
		res.Status(http.StatusCreated)

		svc = &v1.Service{
			Id:   sid,
			Name: "r1",
			Host: "new.example.com",
			Path: "/bar-new",
		}
		res = c.PUT("/v1/services/" + sid).
			WithJSON(svc).
			Expect()
		res.Status(http.StatusOK)

		res = c.GET("/v1/services/" + sid).Expect()
		res.Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.Value("host").Equal("new.example.com")
		body.Value("path").Equal("/bar-new")
	})
	t.Run("upsert service without id fails", func(t *testing.T) {
		svc := goodService()
		res := c.PUT("/v1/services/").
			WithJSON(svc).
			Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " '' is not a valid uuid")
	})
	t.Run("upsert service with not existing CA cert fails", func(t *testing.T) {
		svc := goodService()
		svc.Name = "with-cert"
		svc.Protocol = "https"
		svc.CaCertificates = []string{uuid.NewString()}
		res := c.PUT("/v1/services/" + uuid.NewString()).
			WithJSON(svc).
			Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "data constraint error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0)
		err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.
			String())
		err.Object().ValueEqual("field", "ca_certificates")
	})
	t.Run("upsert a service referencing multiple CA certs successfully", func(t *testing.T) {
		// create CA certificates
		res := c.POST("/v1/ca-certificates").WithJSON(&v1.CACertificate{
			Cert: goodCACertOne,
		}).Expect()
		res.Status(http.StatusCreated)
		body := res.JSON().Path("$.item").Object()
		caCertIDOne := body.Value("id").String().Raw()

		// upsert service
		service := goodService()
		service.Name = "with-good-certs"
		service.Protocol = "https"
		service.CaCertificates = []string{caCertIDOne}
		res = c.POST("/v1/services").WithJSON(service).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body = res.JSON().Path("$.item").Object()
		svcID := body.Value("id").String().Raw()
		body.Value("ca_certificates").Array().Length().Equal(1)
		gotCACertID := body.Value("ca_certificates").Array().Element(0).String().Raw()
		require.True(t, caCertIDOne == gotCACertID)

		// create new CA cert
		res = c.POST("/v1/ca-certificates").WithJSON(&v1.CACertificate{
			Cert: goodCACertTwo,
		}).Expect()
		res.Status(http.StatusCreated)
		body = res.JSON().Path("$.item").Object()
		caCertIDTwo := body.Value("id").String().Raw()

		// upsert service
		service.CaCertificates = []string{caCertIDOne, caCertIDTwo}
		res = c.PUT("/v1/services/" + svcID).WithJSON(service).Expect()
		res.Status(http.StatusOK)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body = res.JSON().Path("$.item").Object()
		body.Value("ca_certificates").Array().Length().Equal(2)
		gotCACertIDOne := body.Value("ca_certificates").Array().Element(0).String().Raw()
		gotCACertIDTwo := body.Value("ca_certificates").Array().Element(1).String().Raw()
		require.True(t, caCertIDOne == gotCACertIDOne)
		require.True(t, caCertIDTwo == gotCACertIDTwo)
	})
}

func TestServiceDelete(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	svc := goodService()
	res := c.POST("/v1/services").WithJSON(svc).Expect()
	id := res.JSON().Path("$.item.id").String().Raw()
	res.Status(http.StatusCreated)
	t.Run("deleting a non-existent service returns 404", func(t *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.DELETE("/v1/services/" + randomID).Expect().Status(http.StatusNotFound)
	})
	t.Run("deleting a service return 204", func(t *testing.T) {
		c.DELETE("/v1/services/" + id).Expect().Status(http.StatusNoContent)
	})
	t.Run("delete request without an ID returns 400", func(t *testing.T) {
		res := c.DELETE("/v1/services/").Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " '' is not a valid uuid")
	})
	t.Run("delete request with an invalid ID returns 400", func(t *testing.T) {
		res := c.DELETE("/v1/services/" + "Not-Valid").Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " 'Not-Valid' is not a valid uuid")
	})
	t.Run("deletes a CA certificate referenced in a service", func(t *testing.T) {
		// create CA certificates
		res := c.POST("/v1/ca-certificates").WithJSON(&v1.CACertificate{
			Cert: goodCACertOne,
		}).Expect()
		res.Status(http.StatusCreated)
		body := res.JSON().Path("$.item").Object()
		caCertID := body.Value("id").String().Raw()

		// upsert service
		service := goodService()
		service.Protocol = "https"
		service.CaCertificates = []string{caCertID}
		res = c.POST("/v1/services").WithJSON(service).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body = res.JSON().Path("$.item").Object()
		body.Value("ca_certificates").Array().Length().Equal(1)
		gotCACertID := body.Value("ca_certificates").Array().Element(0).String().Raw()
		require.True(t, caCertID == gotCACertID)

		// delete CA certificates
		res = c.DELETE("/v1/ca-certificates/" + caCertID).Expect()
		res.Status(http.StatusNoContent)
	})
	t.Run("deleting a service deletes its plugins, routes and plugin on the route", func(t *testing.T) {
		svc := goodService()
		svc.Name = "bar"
		res := c.POST("/v1/services").WithJSON(svc).Expect()
		res.Status(http.StatusCreated)
		sid := res.JSON().Path("$.item.id").String().Raw()

		route := goodRoute()
		route.Service = &v1.Service{Id: sid}
		res = c.POST("/v1/routes").WithJSON(route).Expect()
		res.Status(http.StatusCreated)
		rid := res.JSON().Path("$.item.id").String().Raw()

		plugin := goodKeyAuthPlugin()
		plugin.Service = &v1.Service{Id: sid}
		res = c.POST("/v1/plugins").WithJSON(plugin).Expect()
		res.Status(http.StatusCreated)
		servicePluginID := res.JSON().Path("$.item.id").String().Raw()

		plugin = goodKeyAuthPlugin()
		plugin.Route = &v1.Route{Id: rid}
		res = c.POST("/v1/plugins").WithJSON(plugin).Expect()
		res.Status(http.StatusCreated)
		routePluginID := res.JSON().Path("$.item.id").String().Raw()

		c.DELETE("/v1/services/" + sid).Expect().Status(http.StatusNoContent)
		c.GET("/v1/routes/" + rid).Expect().Status(http.StatusNotFound)
		c.GET("/v1/plugins/" + servicePluginID).Expect().Status(http.StatusNotFound)
		c.GET("/v1/plugins/" + routePluginID).Expect().Status(http.StatusNotFound)
	})
}

func TestServiceRead(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	svc := goodService()
	res := c.POST("/v1/services").WithJSON(svc).Expect()
	id := res.JSON().Path("$.item.id").String().Raw()
	res.Status(http.StatusCreated)
	t.Run("reading a non-existent service returns 404", func(t *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.GET("/v1/services/" + randomID).Expect().Status(http.StatusNotFound)
	})
	t.Run("reading a service return 200", func(t *testing.T) {
		res := c.GET("/v1/services/" + id).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		validateGoodService(body)
	})
	t.Run("reading a service with service name return 200", func(t *testing.T) {
		res := c.GET("/v1/services/" + svc.Name).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		validateGoodService(body)
	})
	t.Run("read request without an ID returns 400", func(t *testing.T) {
		res := c.GET("/v1/services/").Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "required ID is missing")
	})
	t.Run("read request with no name match returns 404", func(t *testing.T) {
		res := c.GET("/v1/services/somename").Expect()
		res.Status(http.StatusNotFound)
	})
	t.Run("read request with invalid name or ID match returns 400", func(t *testing.T) {
		invalidKey := "234wabc?!@"
		res := c.GET("/v1/services/" + invalidKey).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", fmt.Sprintf("invalid ID:'%s'", invalidKey))
	})
	t.Run("read request with very long name or ID match returns 400", func(t *testing.T) {
		longID := strings.Repeat("0123456789", 13)

		res = c.GET("/v1/services/" + longID).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", fmt.Sprintf("invalid ID:'%s'", longID))
	})
}

func TestServiceList(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	svc := goodService()
	res := c.POST("/v1/services").WithJSON(svc).Expect()
	id1 := res.JSON().Path("$.item.id").String().Raw()
	res.Status(http.StatusCreated)
	svc = &v1.Service{
		Name: "bar",
		Host: "bar.com",
		Path: "/bar",
	}
	res = c.POST("/v1/services").WithJSON(svc).Expect()
	id2 := res.JSON().Path("$.item.id").String().Raw()
	res.Status(http.StatusCreated)

	t.Run("list returns multiple services", func(t *testing.T) {
		body := c.GET("/v1/services").Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(2)
		var gotIDs []string
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, []string{id1, id2}, gotIDs)
	})
}

func TestServiceListPagination(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	// Create ten services
	svcName := "myservice-%d"
	svc := goodService()
	for i := 0; i < 10; i++ {
		svc.Name = fmt.Sprintf(svcName, i)
		res := c.POST("/v1/services").WithJSON(svc).Expect()
		res.JSON().Path("$.item.id").String()
		res.Status(http.StatusCreated)
	}
	var tailID string
	var headID string
	body := c.GET("/v1/services").Expect().Status(http.StatusOK).JSON().Object()
	items := body.Value("items").Array()
	items.Length().Equal(10)
	// Get the head's id so that we can make sure that it is consistent
	headID = items.Element(0).Object().Value("id").String().Raw()
	require.NotEmpty(t, headID)
	// Get the tail's id so that we can make sure that it is consistent
	tailID = items.Element(9).Object().Value("id").String().Raw()
	require.NotEmpty(t, tailID)
	body.Value("page").Object().Value("total_count").Number().Equal(10)
	body.Value("page").Object().NotContainsKey("next_page_num")

	t.Run("list size 1 page 1 returns 1 service total_count=10", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("page.size", "1").
			WithQuery("page.number", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(1)

		body.Value("page").Object().Value("total_count").Number().Equal(10)
		body.Value("page").Object().Value("next_page_num").Number().Equal(2)
		firstID := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, firstID)
		require.Equal(t, headID, firstID)
		// Go to last page and get the last element
		body = c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("page.size", "1").
			WithQuery("page.number", "10").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(1)
		body.Value("page").Object().Value("total_count").Number().Equal(10)
		body.Value("page").Object().NotContainsKey("next_page_num")

		lastID := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, lastID)
		require.Equal(t, tailID, lastID)
	})
	t.Run("list page_size 2 and page 1 returns 2 services with total_count=10", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("page.size", "2").
			WithQuery("page.number", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(2)
		body.Value("page").Object().Value("total_count").Number().Equal(10)
		body.Value("page").Object().Value("next_page_num").Number().Equal(2)

		firstID := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, firstID)
		require.Equal(t, headID, firstID)
		// Go to last page and get the last element
		body = c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("page.size", "2").
			WithQuery("page.number", "5").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(2)
		body.Value("page").Object().Value("total_count").Number().Equal(10)
		body.Value("page").Object().NotContainsKey("next_page_num")
		lastID := items.Element(1).Object().Value("id").String().Raw()
		require.NotEmpty(t, lastID)
		require.Equal(t, tailID, lastID)
	})
	t.Run("list page_size 3 and page 1 returns 3 services with total_count=10", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("page.size", "3").
			WithQuery("page.number", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(3)
		body.Value("page").Object().Value("total_count").Number().Equal(10)
		body.Value("page").Object().Value("next_page_num").Number().Equal(2)
		firstID := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, firstID)
		require.Equal(t, headID, firstID)
		// Go to last page and get the last element
		body = c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("page.size", "3").
			WithQuery("page.number", "4").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(1)
		body.Value("page").Object().Value("total_count").Number().Equal(10)
		body.Value("page").Object().NotContainsKey("next_page_num")
		lastID := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, lastID)
		require.Equal(t, tailID, lastID)
	})
	t.Run("list page_size 4 and page 1 returns 4 services with total_count=10", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("page.size", "4").
			WithQuery("page.number", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(4)
		body.Value("page").Object().Value("total_count").Number().Equal(10)
		body.Value("page").Object().Value("next_page_num").Number().Equal(2)
		firstID := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, firstID)
		require.Equal(t, headID, firstID)
		// Go to last page and get the last element
		body = c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("page.size", "4").
			WithQuery("page.number", "3").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(2)
		body.Value("page").Object().Value("total_count").Number().Equal(10)
		body.Value("page").Object().NotContainsKey("next_page_num")
		lastID := items.Element(1).Object().Value("id").String().Raw()
		require.NotEmpty(t, lastID)
		require.Equal(t, tailID, lastID)
	})
	t.Run("list page_size 10 and page 1 returns 10 services with total_count=10", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("page.size", "10").
			WithQuery("page.number", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(10)
		body.Value("page").Object().Value("total_count").Number().Equal(10)
		body.Value("page").Object().NotContainsKey("next_page_num")
		firstID := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, firstID)
		require.Equal(t, headID, firstID)

		lastID := items.Element(9).Object().Value("id").String().Raw()
		require.NotEmpty(t, lastID)
		require.Equal(t, tailID, lastID)
	})
	t.Run("list page_size 10 and page 10 returns no services", func(t *testing.T) {
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("page.size", "10").
			WithQuery("page.number", "10").
			Expect().Status(http.StatusOK).JSON().Object()
		body.NotContainsKey("items")
	})
	t.Run("list page_size 10 and no Page returns 10 services with total_count=10", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("page.size", "10").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(10)
		body.Value("page").Object().Value("total_count").Number().Equal(10)
		body.Value("page").Object().NotContainsKey("next_page_num")
		firstID := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, firstID)
		require.Equal(t, headID, firstID)

		lastID := items.Element(9).Object().Value("id").String().Raw()
		require.NotEmpty(t, lastID)
		require.Equal(t, tailID, lastID)
	})
	t.Run("list no page_size and Page 1 returns 10 services with total_count=10", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("page.number", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(10)
		body.Value("page").Object().Value("total_count").Number().Equal(10)
		body.Value("page").Object().NotContainsKey("next_page_num")
		firstID := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, firstID)
		require.Equal(t, headID, firstID)

		lastID := items.Element(9).Object().Value("id").String().Raw()
		require.NotEmpty(t, lastID)
		require.Equal(t, tailID, lastID)
	})
	t.Run("list page_size 11 and page 1 returns 10 services with total_count=10", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("page.size", "11").
			WithQuery("page.number", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(10)
		body.Value("page").Object().Value("total_count").Number().Equal(10)
		body.Value("page").Object().NotContainsKey("next_page_num")
		firstID := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, firstID)
		require.Equal(t, headID, firstID)

		lastID := items.Element(9).Object().Value("id").String().Raw()
		require.NotEmpty(t, lastID)
		require.Equal(t, tailID, lastID)
	})
	t.Run("list > 1001 page size and page 1 returns error", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("page.size", "1001").
			WithQuery("page.number", "1").
			Expect().Status(http.StatusBadRequest).JSON().Object()
		body.Value("code").Number().Equal(3)
		body.Value("message").String().Equal("invalid page_size '1001', must be within range [1 - 1000]")
	})
	t.Run("list page_size 10 and page < 0 returns error", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("page.size", "10").
			WithQuery("page.number", "-1").
			Expect().Status(http.StatusBadRequest).JSON().Object()
		body.Value("code").Number().Equal(3)
		body.Value("message").String().Equal("invalid page number '-1', page must be > 0")
	})
}
