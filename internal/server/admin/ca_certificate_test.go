package admin

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/stretchr/testify/require"
)

var (
	caCertTemplate = x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"kong_clustering"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	goodCert = `
-----BEGIN CERTIFICATE-----
MIICtjCCAZ6gAwIBAgIJAMajhTkQI3TIMA0GCSqGSIb3DQEBCwUAMBAxDjAMBgNV
BAMMBUhlbGxvMB4XDTIyMDIyNzE4MTAzMVoXDTIyMDMyOTE4MTAzMVowEDEOMAwG
A1UEAwwFSGVsbG8wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDhIxLU
6qkYN6/E8wTNagvx+1aui5EwO+ImIA+RZxgnJsDNsg8R1hpABiaYWSNMa5wUs0mc
S9tR/6vaC9VrGdEQXZs94b2qwUqmYdfsHpnLOYqFiZsg3BYTEnua/OWtEhI4LbIL
wsTf2tLT0fBZZn3aNj/dVlt+almPDN8GML9gEio647tFCC1qkHVRaZGxsDZC5IfD
7EiODwp540+CVQXsGaMJQZT2IoNwN96Cyw9h0ayJK2vJNRavBAohGEC13hTbbx1F
eA8cjExmRW31G4J6kz2V+YGlBpXKPNRXO75kd33/IHaKqb35rGcd3OLZRQkMmSoY
VaLzIEHQF+8HZVB7AgMBAAGjEzARMA8GA1UdEwQIMAYBAf8CAQAwDQYJKoZIhvcN
AQELBQADggEBADGmkpZXyUjwPy1kWQpmve71vRCgWgDDc5eyXOzmKZsjYfGBJG4W
3RcbOrbOLvrffjkdt7K7OqHGCc9J4Jp+FG3UC47wkOa/A1NOhPuBzr+YkKPRHMH7
L+9FLfTFdRnPWnZ4CtNW5kOxAlpnrQDTFZll0AkKWqB3MYGpg2ilhtO7txwKHpgX
T2WuO2v5B6TFtibsaYY8uMn/OpouEun0Cbns+KF9TDcqDO95TZSkoTW0bamVQH2T
OhbS5BCFZxy+rJ2BD+gwtWfu3+8t+kiQeXXWVC+0qZahm98LgVKdXyCMxtviGMhh
dHmsRtc2obmt+51SGycNsRZKPrD1WulKwj8=
-----END CERTIFICATE-----
`
	digest = "a239094c44503b6a75071a098d6ef2fdbf1009343f60bbdbb17f52701cd823b1"
)

func TestCACertificateCreate(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	t.Run("creates a valid certificate", func(t *testing.T) {
		cert := createCACert(t, &caCertTemplate)
		res := c.POST("/v1/ca-certificates").WithJSON(&v1.CACertificate{
			Cert: string(cert),
		}).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.Value("cert").String().Equal(string(cert))
	})
	t.Run("creates a valid certificate and check digest", func(t *testing.T) {
		res := c.POST("/v1/ca-certificates").WithJSON(&v1.CACertificate{
			Cert: goodCert,
		}).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.Value("cert").String().Equal(goodCert)
		body.Value("cert_digest").String().Equal(digest)
	})
	t.Run("creating an invalid certificate fails", func(t *testing.T) {
		res := c.POST("/v1/ca-certificates").WithJSON(&v1.CACertificate{
			Cert: "a",
		}).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		errRes := body.Value("details").Array().Element(0)
		errRes.Object().ValueEqual("type",
			v1.ErrorType_ERROR_TYPE_FIELD.String())
		errRes.Object().ValueEqual("field", "cert")
		errRes.Object().ValueEqual("messages", []string{
			"invalid certificate: invalid PEM-encoded value",
		})
	})
	t.Run("creating an invalid CA certificate fails", func(t *testing.T) {
		template := caCertTemplate
		template.BasicConstraintsValid = false
		cert := createCACert(t, &template)
		res := c.POST("/v1/ca-certificates").WithJSON(&v1.CACertificate{
			Cert: string(cert),
		}).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		errRes := body.Value("details").Array().Element(0)
		errRes.Object().ValueEqual("type",
			v1.ErrorType_ERROR_TYPE_FIELD.String())
		errRes.Object().ValueEqual("field", "cert")
		errRes.Object().ValueEqual("messages", []string{
			"certificate does not appear to be a CA because" +
				"it is missing the \"CA\" basic constraint",
		})
	})
	t.Run("creating an expired CA certificate fails", func(t *testing.T) {
		template := caCertTemplate
		template.NotAfter = time.Now().Add(-24 * time.Hour)
		cert := createCACert(t, &template)
		res := c.POST("/v1/ca-certificates").WithJSON(&v1.CACertificate{
			Cert: string(cert),
		}).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		errRes := body.Value("details").Array().Element(0)
		errRes.Object().ValueEqual("type",
			v1.ErrorType_ERROR_TYPE_FIELD.String())
		errRes.Object().ValueEqual("field", "cert")
		errRes.Object().ValueEqual("messages", []string{
			"certificate expired, \"Not After\" time is in the past",
		})
	})
	t.Run("creating multiple CA certificates fails", func(t *testing.T) {
		cert := string(createCACert(t, &caCertTemplate)) + goodCert
		res := c.POST("/v1/ca-certificates").WithJSON(&v1.CACertificate{
			Cert: cert,
		}).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		errRes := body.Value("details").Array().Element(0)
		errRes.Object().ValueEqual("type",
			v1.ErrorType_ERROR_TYPE_FIELD.String())
		errRes.Object().ValueEqual("field", "cert")
		errRes.Object().ValueEqual("messages", []string{
			"please submit only one certificate at a time",
		})
	})
}

func TestCACertificateUpsert(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	t.Run("upserts a valid certificate", func(t *testing.T) {
		id := uuid.NewString()
		resource := &v1.CACertificate{
			Cert: goodCert,
		}
		res := c.PUT("/v1/ca-certificates/{id}", id).WithJSON(resource).Expect()
		res.Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.Value("id").String().Equal(id)
		body.Value("cert").String().Equal(goodCert)
		body.Value("cert_digest").String().Equal(digest)
		resource.Id = body.Value("id").String().Raw()
	})
	t.Run("upserts a valid certificate ignore digest", func(t *testing.T) {
		id := uuid.NewString()
		resource := &v1.CACertificate{
			Cert:       goodCert,
			CertDigest: "a",
		}
		res := c.PUT("/v1/ca-certificates/{id}", id).WithJSON(resource).Expect()
		res.Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.Value("id").String().Equal(id)
		body.Value("cert").String().Equal(goodCert)
		body.Value("cert_digest").String().Equal(digest)
		resource.Id = body.Value("id").String().Raw()
	})
	t.Run("upsert an existing certificate succeeds", func(t *testing.T) {
		cert := createCACert(t, &caCertTemplate)
		res := c.POST("/v1/ca-certificates").WithJSON(&v1.Certificate{
			Cert: string(cert),
		}).Expect()
		res.Status(201)
		body := res.JSON().Path("$.item").Object()
		id := body.Value("id").String().Raw()

		cert = createCACert(t, &caCertTemplate)
		certUpdate := &v1.CACertificate{
			Id:   id,
			Cert: string(cert),
		}
		res = c.PUT("/v1/ca-certificates/{id}", id).WithJSON(certUpdate).Expect()
		res.Status(http.StatusOK)
		body = res.JSON().Path("$.item").Object()
		body.Value("id").String().Equal(id)
		body.Value("cert").String().Equal(string(cert))
	})
	t.Run("upsert certificate without id fails", func(t *testing.T) {
		res := c.PUT("/v1/ca-certificates/").
			WithJSON(&v1.CACertificate{}).
			Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " '' is not a valid uuid")
	})
	t.Run("upsert an invalid certificate fails", func(t *testing.T) {
		res := c.PUT("/v1/ca-certificates/{id}", uuid.NewString()).WithJSON(&v1.CACertificate{
			Cert: "a",
		}).Expect()
		res.Status(400)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		errRes := body.Value("details").Array().Element(0)
		errRes.Object().ValueEqual("type",
			v1.ErrorType_ERROR_TYPE_FIELD.String())
		errRes.Object().ValueEqual("field", "cert")
		errRes.Object().ValueEqual("messages", []string{
			"invalid certificate: invalid PEM-encoded value",
		})
	})
}

func TestCACertificateRead(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	res := c.POST("/v1/ca-certificates").WithJSON(&v1.CACertificate{
		Cert: goodCert,
	}).Expect()
	res.Status(http.StatusCreated)
	id := res.JSON().Path("$.item.id").String().Raw()
	t.Run("read with an empty id returns 400", func(t *testing.T) {
		res := c.GET("/v1/ca-certificates/").Expect().Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "required ID is missing")
	})
	t.Run("reading a non-existent certificate returns 404", func(t *testing.T) {
		c.GET("/v1/ca-certificates/{id}", uuid.NewString()).
			Expect().Status(http.StatusNotFound)
	})
	t.Run("reading with an exisiting certificate by returns 200", func(t *testing.T) {
		res := c.GET("/v1/ca-certificates/{id}", id).
			Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.Value("id").Equal(id)
		body.Value("cert").Equal(goodCert)
		body.Value("cert_digest").Equal(digest)
	})
}

func TestCACertificateDelete(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	cert := createCACert(t, &caCertTemplate)
	res := c.POST("/v1/ca-certificates").WithJSON(&v1.CACertificate{
		Cert: string(cert),
	}).Expect().Status(http.StatusCreated)
	id := res.JSON().Path("$.item.id").String().Raw()
	t.Run("deleting a non-existent certificate returns 404", func(t *testing.T) {
		dres := c.DELETE("/v1/ca-certificates/{id}", uuid.NewString()).Expect()
		dres.Status(http.StatusNotFound)
	})
	t.Run("delete with an invalid id returns 400", func(t *testing.T) {
		dres := c.DELETE("/v1/ca-certificates/").Expect()
		dres.Status(http.StatusBadRequest)
	})
	t.Run("delete an existing certificate succeeds", func(t *testing.T) {
		dres := c.DELETE("/v1/ca-certificates/{id}", id).Expect()
		dres.Status(http.StatusNoContent)
	})
}

func TestCACertificateList(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	ids := make([]string, 0, 4)
	for i := 0; i < 4; i++ {
		cert := createCACert(t, &caCertTemplate)
		res := c.POST("/v1/ca-certificates").WithJSON(&v1.Certificate{
			Cert: string(cert),
		}).Expect().Status(http.StatusCreated)
		ids = append(ids, res.JSON().Path("$.item.id").String().Raw())
	}

	t.Run("list returns multiple certificates", func(t *testing.T) {
		body := c.GET("/v1/ca-certificates").Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(4)
		gotIDs := make([]string, 0, 4)
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, ids, gotIDs)
	})
	t.Run("list returns multiple certificates with paging", func(t *testing.T) {
		body := c.GET("/v1/ca-certificates").
			WithQuery("page.size", "2").
			WithQuery("page.number", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(2)
		gotIDs := make([]string, 0, 4)
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		body.Value("page").Object().Value("total_count").Number().Equal(4)
		next := body.Value("page").Object().Value("next_page_num").Number().Equal(2).Raw()
		body = c.GET("/v1/ca-certificates").
			WithQuery("page.size", "2").
			WithQuery("page.number", next).
			Expect().Status(http.StatusOK).JSON().Object()
		body.Value("page").Object().Value("total_count").Number().Equal(4)
		body.Value("page").Object().NotContainsKey("next_page")
		items = body.Value("items").Array()
		items.Length().Equal(2)
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, ids, gotIDs)
	})
}

func createCACert(t *testing.T, template *x509.Certificate) (certPEM []byte) {
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	require.Nil(t, err)
	der, err := x509.CreateCertificate(rand.Reader, template, template, &caPrivKey.PublicKey, caPrivKey)
	require.Nil(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}
