package admin

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/stretchr/testify/require"
)

func TestSNICreate(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	t.Run("creating an SNI with a non-existent certificate fails", func(t *testing.T) {
		res := c.POST("/v1/snis").WithJSON(&v1.SNI{
			Name: "example.com",
			Certificate: &v1.Certificate{
				Id: uuid.NewString(),
			},
		}).Expect().Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "data constraint error")
		body.Value("details").Array().Length().Equal(1)
		errRes := body.Value("details").Array().Element(0)
		errRes.Object().ValueEqual("type",
			v1.ErrorType_ERROR_TYPE_REFERENCE.String())
		errRes.Object().ValueEqual("field", "certificate.id")
	})
	t.Run("creates a valid SNI", func(t *testing.T) {
		cert, key := createCert(t, true)
		res := c.POST("/v1/certificates").WithJSON(&v1.Certificate{
			Cert: string(cert),
			Key:  string(key),
		}).Expect().Status(http.StatusCreated)
		certID := res.JSON().Path("$.item.id").String().Raw()
		res = c.POST("/v1/snis").WithJSON(&v1.SNI{
			Name: "*.example.com",
			Certificate: &v1.Certificate{
				Id: certID,
			},
		}).Expect().Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.Value("name").String().Equal("*.example.com")
	})
	t.Run("creating an SNI with an existing name/hostname fails", func(t *testing.T) {
		cert, key := createCert(t, true)
		res := c.POST("/v1/certificates").WithJSON(&v1.Certificate{
			Cert: string(cert),
			Key:  string(key),
		}).Expect().Status(http.StatusCreated)
		certID := res.JSON().Path("$.item.id").String().Raw()
		res = c.POST("/v1/snis").WithJSON(&v1.SNI{
			Name: "*.example.com",
			Certificate: &v1.Certificate{
				Id: certID,
			},
		}).Expect().Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "data constraint error")
		resErr := body.Value("details").Array().Element(0)
		resErr.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.String())
		resErr.Object().ValueEqual("messages", []string{
			"unique-name (type: unique) constraint failed for value '*.example.com': ",
		})
	})
}

func TestSNIUpsert(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	t.Run("upsert a valid SNI", func(t *testing.T) {
		cert, key := createCert(t, true)
		res := c.POST("/v1/certificates").WithJSON(&v1.Certificate{
			Cert: string(cert),
			Key:  string(key),
		}).Expect().Status(http.StatusCreated)
		certID := res.JSON().Path("$.item.id").String().Raw()
		res = c.PUT("/v1/snis/{id}", uuid.NewString()).WithJSON(&v1.SNI{
			Name: "u.example.com",
			Certificate: &v1.Certificate{
				Id: certID,
			},
		}).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.Value("name").String().Equal("u.example.com")
	})
	t.Run("upsert an existing SNI succeeds", func(t *testing.T) {
		cert, key := createCert(t, true)
		res := c.POST("/v1/certificates").WithJSON(&v1.Certificate{
			Cert: string(cert),
			Key:  string(key),
		}).Expect().Status(http.StatusCreated)
		certID := res.JSON().Path("$.item.id").String().Raw()
		res = c.POST("/v1/snis").WithJSON(&v1.SNI{
			Name: "*.test.example.com",
			Certificate: &v1.Certificate{
				Id: certID,
			},
		}).Expect().Status(http.StatusCreated)
		sniID := res.JSON().Path("$.item.id").String().Raw()
		res = c.PUT("/v1/snis/{id}", sniID).WithJSON(&v1.SNI{
			Name: "*.test.example.com",
			Certificate: &v1.Certificate{
				Id: certID,
			},
		}).Expect().Status(http.StatusOK)
		res.JSON().Path("$.item.id").String().Equal(sniID)
	})
	t.Run("upsert an SNI without an id fails", func(t *testing.T) {
		res := c.PUT("/v1/snis/").WithJSON(&v1.SNI{
			Name: "*.u-test.example.com",
			Certificate: &v1.Certificate{
				Id: uuid.NewString(),
			},
		}).Expect().Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " '' is not a valid uuid")
	})
}

func TestSNIRead(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	cert, key := createCert(t, true)
	res := c.POST("/v1/certificates").WithJSON(&v1.Certificate{
		Cert: string(cert),
		Key:  string(key),
	}).Expect().Status(http.StatusCreated)
	certID := res.JSON().Path("$.item.id").String().Raw()
	res = c.POST("/v1/snis").WithJSON(&v1.SNI{
		Name: "example.com",
		Certificate: &v1.Certificate{
			Id: certID,
		},
	}).Expect().Status(http.StatusCreated)
	sniID := res.JSON().Path("$.item.id").String().Raw()
	t.Run("reading with existing SNI id returns 200", func(t *testing.T) {
		res := c.GET("/v1/snis/{id}", sniID).
			Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.Value("id").String().Equal(sniID)
		body.Path("$.certificate.id").String().Equal(certID)
	})
	t.Run("read with an empty id returns 400", func(t *testing.T) {
		res := c.GET("/v1/snis/").Expect().Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "required ID is missing")
	})
	t.Run("reading a non-existent SNI returns 404", func(t *testing.T) {
		c.GET("/v1/snis/{id}", uuid.NewString()).
			Expect().Status(http.StatusNotFound)
	})
}

func TestSNIDelete(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	cert, key := createCert(t, true)
	res := c.POST("/v1/certificates").WithJSON(&v1.Certificate{
		Cert: string(cert),
		Key:  string(key),
	}).Expect().Status(http.StatusCreated)
	certID := res.JSON().Path("$.item.id").String().Raw()
	res = c.POST("/v1/snis").WithJSON(&v1.SNI{
		Name: "example.com",
		Certificate: &v1.Certificate{
			Id: certID,
		},
	}).Expect().Status(http.StatusCreated)
	sniID := res.JSON().Path("$.item.id").String().Raw()
	t.Run("deleting an existing SNI succeeds", func(t *testing.T) {
		c.DELETE("/v1/snis/{id}", sniID).Expect().Status(http.StatusOK)
	})
	t.Run("deleting a non-existent SNI fails", func(t *testing.T) {
		c.DELETE("/v1/snis/{id}", uuid.NewString()).Expect().Status(http.StatusNotFound)
	})
	t.Run("delete with an invalid SNI id fails", func(t *testing.T) {
		res := c.DELETE("/v1/snis/").Expect().Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " '' is not a valid uuid")
	})
}

func TestSNIList(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	cert, key := createCert(t, true)
	res := c.POST("/v1/certificates").WithJSON(&v1.Certificate{
		Cert: string(cert),
		Key:  string(key),
	}).Expect().Status(http.StatusCreated)
	certID := res.JSON().Path("$.item.id").String().Raw()

	ids := make([]string, 0, 4)
	prefixes := []string{"one", "two", "three", "four"}
	for _, prefix := range prefixes {
		res = c.POST("/v1/snis").WithJSON(&v1.SNI{
			Name: fmt.Sprintf("%s.example.com", prefix),
			Certificate: &v1.Certificate{
				Id: certID,
			},
		}).Expect().Status(http.StatusCreated)
		ids = append(ids, res.JSON().Path("$.item.id").String().Raw())
	}
	t.Run("list returns multiple SNIs", func(t *testing.T) {
		body := c.GET("/v1/snis").Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(4)
		gotIDs := make([]string, 0, 4)
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, ids, gotIDs)
	})
	t.Run("list returns multiple SNIs with paging", func(t *testing.T) {
		body := c.GET("/v1/snis").
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
		body = c.GET("/v1/snis").
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
