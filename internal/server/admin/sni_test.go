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

var (
	certOne = `
-----BEGIN CERTIFICATE-----
MIIC5DCCAcygAwIBAgIBATANBgkqhkiG9w0BAQsFADAaMRgwFgYDVQQKDA9rb25n
X2NsdXN0ZXJpbmcwHhcNMjIwMzIwMDg0MjU2WhcNMjIwMzIxMDg0MjU2WjAaMRgw
FgYDVQQKDA9rb25nX2NsdXN0ZXJpbmcwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
ggEKAoIBAQCqDSnpvSXOyA5/mBqGaNYu7ZXnP9tPWe4MyjqxqTT5uDsBQ5jz6yET
pONbZ0cSkf+ek79R8t8TfpCuBE2Fc5dhcz4g31NCaqlJCogddVJ4fpxlU8cTX0Q+
zpv9xRdMvTLBgheYJRDHe6stccHOs8z0dwiN6HS8ZTsucyA243TCoFyPTzelmkj1
rZRMPXtQFPE8uiF+lctrtwQFvuC5MyyLUlHqu6b56Tq0veaUlSYtFtxNSu/o6f8A
cqGszF2+4z897rPynetlpD29qOf2JqfazVJPkVcTJ812gWXTJNTj/m6tokOKUstA
+MYe0XsFCDsZLRChb8xXJHOAxJuYs8GbAgMBAAGjNTAzMA4GA1UdDwEB/wQEAwIF
oDATBgNVHSUEDDAKBggrBgEFBQcDATAMBgNVHRMBAf8EAjAAMA0GCSqGSIb3DQEB
CwUAA4IBAQBb8aHZlMo80HJuHpATbLH1JiN7WwwM1kugsLOpFHBh3gXYlnRF8Qb2
N5JUpL9/leowb/V8FOFkd5SZYHcdrMfHBJSEUItoaPIELBDpQ789c03iEWEe/HTr
6Pg4h5NKllbpBXBXsOaAqyEk+Tm77m4ME0Mrz2FjI5dtvU1S2UsSd8AnB2m/WM8s
haLQqfNDVy/Trj/yCLGbsffxiPlUuZAIcqysUqMfYUKBHQkVJxLdul83RIdqH71q
6RPU9hAVl0jaSU8kSjIfX6C30e9IDKRY+ThZGPhjpiHELAWrovnYpkyYjyuaHslc
KM775mAY9wViKTvMvGaIHuJoc6iGfW3r
-----END CERTIFICATE-----
`

	keyOne = `
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCqDSnpvSXOyA5/
mBqGaNYu7ZXnP9tPWe4MyjqxqTT5uDsBQ5jz6yETpONbZ0cSkf+ek79R8t8TfpCu
BE2Fc5dhcz4g31NCaqlJCogddVJ4fpxlU8cTX0Q+zpv9xRdMvTLBgheYJRDHe6st
ccHOs8z0dwiN6HS8ZTsucyA243TCoFyPTzelmkj1rZRMPXtQFPE8uiF+lctrtwQF
vuC5MyyLUlHqu6b56Tq0veaUlSYtFtxNSu/o6f8AcqGszF2+4z897rPynetlpD29
qOf2JqfazVJPkVcTJ812gWXTJNTj/m6tokOKUstA+MYe0XsFCDsZLRChb8xXJHOA
xJuYs8GbAgMBAAECggEAL2yOZM5ITfvC91iPBS8VlG7T4HMRkXauCckYR1W+HWqA
oiCc9mF7jwPsGCCcVJR86legApWuGrywUqeGixIqhJXkHLzLdnlSjrkuLrD6d/ov
WZ7cpQ1rdeye2k3t6ovVLNxAAkFMBaX6nijceO3x2becnh7W93dv4stej5Atjt0/
e16cnTOGZzex+oWwA1p6co5XXrJ8vwk5i3VvrXZBMJNZ5iuUn3EkgxmZOpKlBlLj
Ze4xi0KysrVmEXICqi2E3aAvgh1gQBnpnvRxt/2Z7/0YPK4B4xnU6GGWnqq7V6vJ
rYuMOm76TSRf7fwLP9tQdoOeaLHNYUC1DyhdFTyeQQKBgQDEUUEH9sg91LHsu9ay
oKV/ZM0RYopLUcMI7JB+y4l+SJgAEDMfvofh+a6tW6Zohhao2ehrVBOOysF6bAre
FfVXG9eB7ka0yFHAPYPQWBBtDaHjU8w+CCEU6TBWrxx2WCrCVjUSpBe+4PLfTckx
lHx9E3C/rIIXd/EKwW0jIep/OwKBgQDdv7bBv0Mqu0PiXG1qZEvwyRT4Dz6ssPky
fDVpdyOn3Krn3zDZb9MU2dg29wTZWA91SN3FHifcFVz9h30L5/TwAL1wUH8gk0qJ
0HVQJHLJbFO6Qi2UZfm8kabZmonKdkTM3D+okpgK2MBJlimlwk8WwdDrz2bV0Jcz
zdi/sYxhIQKBgEa8eFAkTZZp0wpXzE5kr/0tFu7SsL3e4gWPJ6loMUx9X7d2HtWr
U07LJnN0eItk9Tk1+xbhHoLu77PqxierhdEzSP2aG4P8QeigwaQKdzC0HsbIZOld
CH5+X1p8kibaMd4ALfNfiObQKvLnFj11IT34CUInKGDIaOPVOjvUdqgNAoGBAJ8j
rAvlsFVlaXV1MYzuB9X46VSQ1EDpDR4fJ9HVj8AzTG/1rEAP0aOgJ1xi8JbubMGW
FpoVZzO6HS9R4fr+b7kiPtHw4xtEuXSoJtjqH3rQhFIilkVu3chnmx+Fmae0MvH2
irT256i5H15wJtlv1oSVedMR2FJQTYL/ErOXvxAhAoGAR9hiIuRVXSVX36bmg/pW
9baA8DrXFjvDjokM8rJ6eGanX7WyGAcbw2TnHyuzHnfzFhU+Q05E/SPkSacbIAJb
xPMcHozAttjhFmcMrqnxLLesT1d/ppYwteXqgruzYNkWBR0wIngfygMuvr4H2iDV
z2xRZq40uPqd2Bbq6crjhF8=
-----END PRIVATE KEY-----
`

	certTwo = `
-----BEGIN CERTIFICATE-----
MIIC5DCCAcygAwIBAgIBATANBgkqhkiG9w0BAQsFADAaMRgwFgYDVQQKDA9rb25n
X2NsdXN0ZXJpbmcwHhcNMjIwMzIwMDg1MTExWhcNMjIwMzIxMDg1MTExWjAaMRgw
FgYDVQQKDA9rb25nX2NsdXN0ZXJpbmcwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
ggEKAoIBAQDsnytihZ17x5ahLojjfijRyuqLSm0VSGENFhmO+M7GCieBgcC4H/kx
vOstrTKNgRjrIjmhY39F3/SpYcZc6YT7a0nzDuWHcpwaZdYFj/6xlc8CisaDg8t7
+l/0lNGiEDNjLl7z/iTikqsboGjbKmNm/cNFiw7INE+rehHEDLUoLVm0aCocxvvq
HD0kfbi+07ws1l3g0xCAfjNe2IhcNk+AjppxnFmBKDgcoPKz5c1uT/QkCNk7ak2E
+4Txo3GA6VjvJaixt+KFKfT0b/0ayG4dEE2txw94/PP4urfFpRU+8PO0Nn56wNJX
5BjQhTgzJofa4PVRmy4hMU6hgY+xEPw7AgMBAAGjNTAzMA4GA1UdDwEB/wQEAwIF
oDATBgNVHSUEDDAKBggrBgEFBQcDATAMBgNVHRMBAf8EAjAAMA0GCSqGSIb3DQEB
CwUAA4IBAQCV/iw+XvuSlvqUsUxywRb2ZLW2/YwxtJBeMOX33vqgQjGvym2II8V0
8YHiYASIGVN7NOPdAoU7OZ79Xh2i6WgcsBnT+T8jLjcynHXK2vZbQMzGJkAOTWXU
cTfYQvoGiMSreQqfQ2eZ5uv2tobURAfLzudqfGzXRZYb5ZnzPOazfCEyYEveWvNm
pelr8+TobiN2Ab3F5QzXfexD2CZRJaFkVCcUOvetWQy7OCdhx0HlfI0axAV3f0Wf
vleiC6J6es0jhovsuvZqzVShoUkazr6WhnOC5t7WbcagdpWuFVpzmIWWcso/NviQ
Gsf38f0JAIKqmgjM5zyEAawkGYX426Hu
-----END CERTIFICATE-----`

	keyTwo = `
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDsnytihZ17x5ah
LojjfijRyuqLSm0VSGENFhmO+M7GCieBgcC4H/kxvOstrTKNgRjrIjmhY39F3/Sp
YcZc6YT7a0nzDuWHcpwaZdYFj/6xlc8CisaDg8t7+l/0lNGiEDNjLl7z/iTikqsb
oGjbKmNm/cNFiw7INE+rehHEDLUoLVm0aCocxvvqHD0kfbi+07ws1l3g0xCAfjNe
2IhcNk+AjppxnFmBKDgcoPKz5c1uT/QkCNk7ak2E+4Txo3GA6VjvJaixt+KFKfT0
b/0ayG4dEE2txw94/PP4urfFpRU+8PO0Nn56wNJX5BjQhTgzJofa4PVRmy4hMU6h
gY+xEPw7AgMBAAECggEAAv48cUGZbWBn8mABUUdeQtEbSGnHmXZR6/V0m09gZjbo
qwW2J14YK93k564CLrIMW6USL41vpbWghaf7917o1LlVtSJiGuWDPf49x9I7eYmY
lcKlojI/l7DiF9juEeu8iquifdmgI9GRIodT7DnMChh5qN6KcFPhEh04Lk+u1vQ9
rNF5o4MVo+AmdjAhQpyUN760//S8DhmeuMsJ+Vu5nC15eKrazFPwUdKgrmiVibVF
KCYZnb1wpknqFxn/d/nCY1x32r9gzdk/xcOGFL0ZAv8cAdN93D/9of+ppXr1cJW2
s9PrOOxoeKSAfnabxgxLznGoSmcIBV6x2/CmC3XgAQKBgQD70EUwT8eQQjQUEuvA
VhI+pwfEx3NkdXBiKQw6F3YhEjbf4h0b0rkjTN0/LjDNvoNjhwFZSfdjuL5ocDh3
P//rpam6jTy3ApBsi0S5kI9X9ORjvmoL3M84xQ7+logrlsWHvOERJLygs93CfmEl
0sT6/HyZnLgtKNzzrSI9/GasOwKBgQDwjj4Cc4eOkDGNU3jRlFjbm8Fl8QCOcum9
78XkiMfpSeCjWzVK51XeOgKKv644ub15oeQPR0ePZsa4wZ9MZ3EwakMLGocLA/kS
IlxpcSePzBwp4LIIUmuHU2L0bjszSyvzXVtHZs5W/ZYy0aVjY8GiQa8r0r3Q+zJQ
4HO0fWnwAQKBgHoCPaPc4+rHyQf46vV0Pr7Qm3kC0qxYIq3NCbmT6I65jpEHs+bp
QP8TnRehv8/QgUTWAxdKOW2987QSu6k7/zokOIrFKCfcPDH7gL8QhgOuCoMxnZxF
zrnI8Sz1ruC/2tGb+Mkfra2HuOkl5tg2uW6Kq6yaPLrU08nVl4PFKdJFAoGBAOLO
SlsHUHUzMPU+EXkQ9KLCfRs/mrW0VPw3OQ9bg5lKhZmf4mRoL0bizQjC52ImhiZL
ZHqfSzJCxfTm4eoo0cjDN8kdTtws98aITTdBb/qdiKRXbaR5CVdDYNZzpC+dnafM
isaMgNn2KIprrhNCuAvjAGVCQqPqY6trpMw1PxABAoGAesmIpFrVYpaj/uCuQnAs
5YZ83kThrSqGuKMmON8RoXMTNQ/VI4dSnmar8SZPG2L8UBLsdXRYCuaEKFs5pfzN
y3qHMGliWPMN9FV60VJRpjw6/d8zEr9wFde+ArydE00krpqNoDukfkb+db5QrOKF
mpEidaFIHN/a8TrYUkCPmxs=
-----END PRIVATE KEY-----`
)

func TestSNICreate(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	// create certificate
	res := c.POST("/v1/certificates").WithJSON(&v1.Certificate{
		Cert: certOne,
		Key:  keyOne,
	}).Expect().Status(http.StatusCreated)
	certID := res.JSON().Path("$.item.id").String().Raw()

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
			"name (type: unique) constraint failed for value '*.example.com': ",
		})
	})
}

func TestSNIUpsert(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	// create certificate
	res := c.POST("/v1/certificates").WithJSON(&v1.Certificate{
		Cert: certOne,
		Key:  keyOne,
	}).Expect().Status(http.StatusCreated)
	certID := res.JSON().Path("$.item.id").String().Raw()

	t.Run("upsert a valid SNI", func(t *testing.T) {
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
		res = c.POST("/v1/snis").WithJSON(&v1.SNI{
			Name: "*.test.example.com",
			Certificate: &v1.Certificate{
				Id: certID,
			},
		}).Expect().Status(http.StatusCreated)
		sniID := res.JSON().Path("$.item.id").String().Raw()
		res = c.PUT("/v1/snis/{id}", sniID).WithJSON(&v1.SNI{
			Name: "*.test-up-name.example.com",
			Certificate: &v1.Certificate{
				Id: certID,
			},
		}).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.Value("id").String().Equal(sniID)
		body.Value("name").String().Equal("*.test-up-name.example.com")
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

	// create certificate
	res := c.POST("/v1/certificates").WithJSON(&v1.Certificate{
		Cert: certOne,
		Key:  keyOne,
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
		body.Value("name").String().Equal("example.com")
		body.Path("$.certificate.id").String().Equal(certID)
	})
	t.Run("reading with existing SNI name returns 200", func(t *testing.T) {
		res := c.GET("/v1/snis/" + "example.com").
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
	t.Run("read SNI with no name match returns 404", func(t *testing.T) {
		res := c.GET("/v1/snis/somename").Expect()
		res.Status(http.StatusNotFound)
	})
}

func TestSNIDelete(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	// create certificate
	res := c.POST("/v1/certificates").WithJSON(&v1.Certificate{
		Cert: certOne,
		Key:  keyOne,
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
		c.DELETE("/v1/snis/{id}", sniID).Expect().Status(http.StatusNoContent)
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

	// create certificates
	res := c.POST("/v1/certificates").WithJSON(&v1.Certificate{
		Cert: certOne,
		Key:  keyOne,
	}).Expect().Status(http.StatusCreated)
	certIDOne := res.JSON().Path("$.item.id").String().Raw()

	res = c.POST("/v1/certificates").WithJSON(&v1.Certificate{
		Cert: certTwo,
		Key:  keyTwo,
	}).Expect().Status(http.StatusCreated)
	certIDTwo := res.JSON().Path("$.item.id").String().Raw()

	idsCertOne := make([]string, 0, 4)
	prefixes := []string{"one", "two", "three", "four"}
	for _, prefix := range prefixes {
		res = c.POST("/v1/snis").WithJSON(&v1.SNI{
			Name: fmt.Sprintf("%s.example.com", prefix),
			Certificate: &v1.Certificate{
				Id: certIDOne,
			},
		}).Expect().Status(http.StatusCreated)
		idsCertOne = append(idsCertOne, res.JSON().Path("$.item.id").String().Raw())
	}
	idsCertTwo := make([]string, 0, 4)
	for _, prefix := range prefixes {
		res = c.POST("/v1/snis").WithJSON(&v1.SNI{
			Name: fmt.Sprintf("%s.new.example.com", prefix),
			Certificate: &v1.Certificate{
				Id: certIDTwo,
			},
		}).Expect().Status(http.StatusCreated)
		idsCertTwo = append(idsCertTwo, res.JSON().Path("$.item.id").String().Raw())
	}
	ids := make([]string, 0, 8)
	ids = append(ids, idsCertOne...)
	ids = append(ids, idsCertTwo...)
	t.Run("list returns multiple SNIs", func(t *testing.T) {
		body := c.GET("/v1/snis").Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(8)
		gotIDs := make([]string, 0, 8)
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, ids, gotIDs)
	})
	t.Run("list returns multiple SNIs with paging", func(t *testing.T) {
		body := c.GET("/v1/snis").
			WithQuery("page.size", "4").
			WithQuery("page.number", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(4)
		gotIDs := make([]string, 0, 8)
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		body.Value("page").Object().Value("total_count").Number().Equal(8)
		next := body.Value("page").Object().Value("next_page_num").Number().Equal(2).Raw()
		body = c.GET("/v1/snis").
			WithQuery("page.size", "4").
			WithQuery("page.number", next).
			Expect().Status(http.StatusOK).JSON().Object()
		body.Value("page").Object().Value("total_count").Number().Equal(8)
		body.Value("page").Object().NotContainsKey("next_page")
		items = body.Value("items").Array()
		items.Length().Equal(4)
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, ids, gotIDs)
	})
	t.Run("list snis by certificate", func(t *testing.T) {
		body := c.GET("/v1/snis").WithQuery("certificate_id", certIDOne).
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(4)
		var gotIDs []string
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, idsCertOne, gotIDs)

		body = c.GET("/v1/snis").WithQuery("certificate_id", certIDTwo).
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(4)
		gotIDs = nil
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, idsCertTwo, gotIDs)
	})
	t.Run("list snis by certificate with paging", func(t *testing.T) {
		body := c.GET("/v1/snis").
			WithQuery("certificate_id", certIDOne).
			WithQuery("page.size", "2").
			WithQuery("page.number", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(2)
		body.Value("page").Object().Value("total_count").Number().Equal(4)
		body.Value("page").Object().Value("next_page_num").Number().Equal(2)
		id1Got := items.Element(0).Object().Value("id").String().Raw()
		id2Got := items.Element(1).Object().Value("id").String().Raw()
		// Next
		body = c.GET("/v1/snis").
			WithQuery("certificate_id", certIDOne).
			WithQuery("page.size", "2").
			WithQuery("page.number", "2").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(2)
		body.Value("page").Object().Value("total_count").Number().Equal(4)
		body.Value("page").Object().NotContainsKey("next_page_num")
		id3Got := items.Element(0).Object().Value("id").String().Raw()
		id4Got := items.Element(1).Object().Value("id").String().Raw()
		require.ElementsMatch(t, idsCertOne, []string{id1Got, id2Got, id3Got, id4Got})
	})
	t.Run("list snis by certificate - no sni associated with certificate", func(t *testing.T) {
		body := c.GET("/v1/snis").WithQuery("certificate_id", uuid.NewString()).
			Expect().Status(http.StatusOK).JSON().Object()
		body.Empty()
	})
	t.Run("list snis by certificate - invalid certificate UUID", func(t *testing.T) {
		body := c.GET("/v1/snis").WithQuery("certificate_id", "invalid-uuid").
			Expect().Status(http.StatusBadRequest).JSON().Object()
		body.Keys().Length().Equal(2)
		body.ValueEqual("code", 3)
		body.ValueEqual("message", "certificate_id 'invalid-uuid' is not a UUID")
	})
}
