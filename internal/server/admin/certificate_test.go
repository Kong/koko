package admin

import (
	"crypto/ecdsa"
	"crypto/elliptic"
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

var certTemplate = x509.Certificate{
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

func TestCertificateCreate(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	t.Run("creates a valid certificate", func(t *testing.T) {
		cert, key := createCert(t, true)
		res := c.POST("/v1/certificates").WithJSON(&v1.Certificate{
			Cert: string(cert),
			Key:  string(key),
		}).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.Value("cert").String().Equal(string(cert))
		body.Value("key").String().Equal(string(key))
	})
	t.Run("creating an invalid certificate fails", func(t *testing.T) {
		res := c.POST("/v1/certificates").WithJSON(&v1.Certificate{
			Cert: "a",
			Key:  "b",
		}).Expect()
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
		require.ElementsMatch(t, []string{"cert", "key"}, fields)
	})
}

func TestCertificateUpsert(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	t.Run("upserts a valid certificate", func(t *testing.T) {
		id := uuid.NewString()
		cert, key := createCert(t, true)
		resource := &v1.Certificate{
			Cert: string(cert),
			Key:  string(key),
		}
		res := c.PUT("/v1/certificates/{id}", id).WithJSON(resource).Expect()
		res.Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.Value("id").String().Equal(id)
		body.Value("cert").String().Equal(string(cert))
		body.Value("key").String().Equal(string(key))
		resource.Id = body.Value("id").String().Raw()
	})
	t.Run("upsert an existing certificate succeeds", func(t *testing.T) {
		cert, key := createCert(t, true)
		res := c.POST("/v1/certificates").WithJSON(&v1.Certificate{
			Cert: string(cert),
			Key:  string(key),
		}).Expect()
		res.Status(201)
		body := res.JSON().Path("$.item").Object()
		id := body.Value("id").String().Raw()

		certAlt, keyAlt := createCert(t, false)
		certUpdate := &v1.Certificate{
			Id:      id,
			Cert:    body.Value("cert").String().Raw(),
			Key:     body.Value("key").String().Raw(),
			CertAlt: string(certAlt),
			KeyAlt:  string(keyAlt),
		}
		res = c.PUT("/v1/certificates/{id}", id).WithJSON(certUpdate).Expect()
		res.Status(http.StatusOK)
		body = res.JSON().Path("$.item").Object()
		body.Value("id").String().Equal(id)
		body.Value("cert").String().Equal(string(cert))
		body.Value("key").String().Equal(string(key))
		body.Value("cert_alt").String().Equal(string(certAlt))
		body.Value("key_alt").String().Equal(string(keyAlt))
	})
	t.Run("upsert certificate without id fails", func(t *testing.T) {
		res := c.PUT("/v1/certificates/").
			WithJSON(&v1.Certificate{}).
			Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " '' is not a valid uuid")
	})
	t.Run("upsert an invalid certificate fails", func(t *testing.T) {
		res := c.PUT("/v1/certificates/{id}", uuid.NewString()).WithJSON(&v1.Certificate{
			Cert: "a",
			Key:  "b",
		}).Expect()
		res.Status(400)
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
		require.ElementsMatch(t, []string{"cert", "key"}, fields)
	})
}

func TestCertificateRead(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	cert, key := createCert(t, true)
	res := c.POST("/v1/certificates").WithJSON(&v1.Certificate{
		Cert: string(cert),
		Key:  string(key),
	}).Expect().Status(http.StatusCreated)
	id := res.JSON().Path("$.item.id").String().Raw()
	t.Run("read with an empty id returns 400", func(t *testing.T) {
		c.GET("/v1/certificates/").Expect().Status(http.StatusBadRequest)
	})
	t.Run("reading a non-existent certificate returns 404", func(t *testing.T) {
		c.GET("/v1/certificates/{id}", uuid.NewString()).
			Expect().Status(http.StatusNotFound)
	})
	t.Run("reading with an exisiting certificate by returns 200", func(t *testing.T) {
		res := c.GET("/v1/certificates/{id}", id).
			Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.Value("id").Equal(id)
		body.Value("cert").Equal(string(cert))
		body.Value("key").Equal(string(key))
	})
}

func TestCertificateDelete(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	cert, key := createCert(t, true)
	res := c.POST("/v1/certificates").WithJSON(&v1.Certificate{
		Cert: string(cert),
		Key:  string(key),
	}).Expect().Status(http.StatusCreated)
	id := res.JSON().Path("$.item.id").String().Raw()
	t.Run("deleting a non-existent certificate returns 404", func(t *testing.T) {
		dres := c.DELETE("/v1/certificates/{id}", uuid.NewString()).Expect()
		dres.Status(http.StatusNotFound)
	})
	t.Run("delete with an invalid id returns 400", func(t *testing.T) {
		dres := c.DELETE("/v1/certificates/").Expect()
		dres.Status(http.StatusBadRequest)
	})
	t.Run("delete an existing certificate succeeds", func(t *testing.T) {
		dres := c.DELETE("/v1/certificates/{id}", id).Expect()
		dres.Status(http.StatusNoContent)
	})
}

func TestCertificateList(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	ids := make([]string, 0, 4)
	for i := 0; i < 4; i++ {
		cert, key := createCert(t, true)
		res := c.POST("/v1/certificates").WithJSON(&v1.Certificate{
			Cert: string(cert),
			Key:  string(key),
		}).Expect().Status(http.StatusCreated)
		ids = append(ids, res.JSON().Path("$.item.id").String().Raw())
	}

	t.Run("list returns multiple certificates", func(t *testing.T) {
		body := c.GET("/v1/certificates").Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(4)
		gotIDs := make([]string, 0, 4)
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, ids, gotIDs)
	})
	t.Run("list returns multiple certificates with paging", func(t *testing.T) {
		body := c.GET("/v1/certificates").
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
		body = c.GET("/v1/certificates").
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

func createCert(t *testing.T, genRSA bool) (certPEM []byte, keyPEM []byte) {
	var prvkey interface{}
	var pubkey interface{}
	if genRSA {
		rsakey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.Nil(t, err)
		require.NotNil(t, rsakey)
		prvkey = rsakey
		pubkey = &rsakey.PublicKey
	} else {
		ecdsakey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.Nil(t, err)
		require.NotNil(t, ecdsakey)
		prvkey = ecdsakey
		pubkey = &ecdsakey.PublicKey
	}

	der, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, pubkey, prvkey)
	require.Nil(t, err)
	require.NotNil(t, der)

	key, err := x509.MarshalPKCS8PrivateKey(prvkey)
	require.Nil(t, err)
	require.NotNil(t, key)

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key})
	return
}
