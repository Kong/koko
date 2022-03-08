package resource

import (
	"testing"

	"github.com/google/uuid"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/stretchr/testify/require"
)

var (
	goodCertOne = `
-----BEGIN CERTIFICATE-----
MIIFEDCCAvigAwIBAgIBATANBgkqhkiG9w0BAQsFADAaMRgwFgYDVQQKDA9rb25n
X2NsdXN0ZXJpbmcwHhcNMjIwMzA4MTIxMzQxWhcNMzIwMjI0MTIxMzQxWjAaMRgw
FgYDVQQKDA9rb25nX2NsdXN0ZXJpbmcwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAw
ggIKAoICAQDoIyEHnkYWxigP4RlFQa/eaIg8TZhcMYvd58cZK36KM3ydN4h5Zjza
SfCQewLFFUMCEppEiD3rPO0feE1m5L+p9SFrUPoG9H98B3WVAuq1DNw0+mPzLu0a
uKB8oQz7TnOZW1DkDNyrhXYHCppFT+xmvMp1pdxCfaZjayU/tY9Ef8G92T8CwBmK
Gu7LTUAyAsY8lz/kaTtXfrhy5ZfMfZX/HkFEUVHPGKfacbBeQOe/kwr88W3NHnS/
J8zYjAUFg0MOI1FS9ImkyOHOIl3/n6buN5mpOTWDklpjURt1BnPqG6WSLn/PMRI7
N0bWX/7hvuoNOR9CQNAt/0aSu+WWco2IdeJK3AJNymiTDRjeH9zz9IqtxWkIvIRU
lXwqKqupj+dNY1vNftyFAGDQrYrZkhouITid6lViDPXi/Pzpb6huD9+u1irPcSe+
uyJ0vEjpXNV0NKl5XZVBeo6fd6WCMGTrWfAqZ5DJrulzPiLZxAci/ZwwgsA1rrdv
+mThAIq1MhkocLjaBfHg+MH9mwm8wzlIEyZfckOb/i/0FdGoLOt3Bpj586n4Tptk
hB4Ns1HyCIdFOt/JDO4WgnxWGHJK1Gm8fZW7sdF8z2Rn+mvKXLyWLQz5Z3hi3HXC
qQJ0fF1Qaw6uVw8325czmtDvCvWUGy4rakODTWZxJ84qerieKiQ75QIDAQABo2Ew
XzAOBgNVHQ8BAf8EBAMCAoQwHQYDVR0lBBYwFAYIKwYBBQUHAwIGCCsGAQUFBwMB
MA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFPrtYHd/q/jd6fwk89I1W3kwbIt5
MA0GCSqGSIb3DQEBCwUAA4ICAQBLjCJpCQBIg/wSMyZ4rU0Iwiw4U+kYg9Z1qLDl
0cZk53RGmyWJgxyY7zJ312qfH3yCXC5Ke7JljNh8f+3StXZHlgtOhygBR6EzQGWY
iC2Yhk44QaDJG4aw/iUQDnBahXgial/dYqdcPLgGJnlxrpmmFtq35b2rxXOoDsAR
rskcjSlvvaFfDcjCbrxLv6p/V/Jd67VBoirWHyuxr22ObWS+ixOGLmTrNeyyLL6K
pEcVpibL8SfpVk9PGkcizUHQBTd0hY8oc7ZiOEa2oSrY6zzOew/vo/PWfCBGsLr4
XD2g2sCLpLJ0/7Viwmnq8EQ/1mdt/oLp/2a1KQjW1a2A2BgHBVagMAp4DypznwuY
DHl7GctIoOBLijTF+iJyghotBUtTjHrHwu5S6kPzr2nJNKLd5y1TmU20iQM5SbCL
bS8yDehYS0ru3MJh8Xg0HfxeKgkqwhfkYzrSKKJ4sIhpg8aw80+ZPon5BsYoP/Mc
LhoRTOEjbTOA5pHf+maJtUi4e0QkAIgJEq6EHnUoitF7pL3QnqLxeGh0f5o+c5p1
tM2qtmm1g+Ostn3kYsCWvf37YUPl1fyKpC11/XcwHd86VSs4FGdSVo0AAhgxLrfR
exZOTeXvHtOd8cIY+yRmQz2IO0ALBTDR57itEVDK8VO70f5JJ3fHV+0KDB6MYpCt
NkbDsw==
-----END CERTIFICATE-----
`
	goodCertTwo = `
-----BEGIN CERTIFICATE-----
MIIFEDCCAvigAwIBAgIBATANBgkqhkiG9w0BAQsFADAaMRgwFgYDVQQKDA9rb25n
X2NsdXN0ZXJpbmcwHhcNMjIwMzA4MTIxNDAxWhcNMzIwMjI0MTIxNDAxWjAaMRgw
FgYDVQQKDA9rb25nX2NsdXN0ZXJpbmcwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAw
ggIKAoICAQCb3IurOghvo1F4QYimcr8GXJ5mwf8DasIptkn4SmbS4fqk+TGoXnup
+vAdl1MRaBu7aaKbMmnG3kv1Wr+mM+Tqz0fTye69y9iPVLnAYPiWmwSBbUsX98YL
PZYDJMKFeOG/0LWEMq4u3OYK8esqXtrIwwcHV2Qiv0dxqzOAinS4R7Cs++/NOETl
2XCga2VFiJxc56NZtEAxvE3oiMeoUEt1iHC6oShm9T2G1ZctKR1/6IPkWgWnoosb
POyMPdEgP0vqBSAtuqvKNoqHYuzfn4mtH+3N3Bq0xu3lpwcOOQDh5qDQ/vzOdKku
9ngBCQNekg3aAl4fGga1FLyOgftxR9IJltu/ZZ5ZCbCWm/zKcuNri3BH/KM/wPgM
29VWJUgU8Q7oiIrt6cTXdBBg16octUlLf45KQd8npoPJKXA6/janXwXFcPg+rSyC
c2R21wvL/OHt6vqIktaIlqvPBzyX8dXEYwenYxRd3AUoRuAtdDhxnYZ7+3wb4xqQ
rjaiHX1YzrebDgxBk9tOWEFsklftIPJtkSjgtVAE6nQ/zi4pa5CAJG7FrVJPPbUu
uzc3WEnjyrjUJ0lTfNblUDkbsXbn2wEW3yibZ567LTttDbaJknCAX+5eo/MoGbgw
koht4d1kQNCZCkRCEqeKDmYYHaPVDmCoO+O7iq9A+CEOu6BSm8VisQIDAQABo2Ew
XzAOBgNVHQ8BAf8EBAMCAoQwHQYDVR0lBBYwFAYIKwYBBQUHAwIGCCsGAQUFBwMB
MA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFG2lQ8uW8kFhoRqPnaK9axNkC4Gt
MA0GCSqGSIb3DQEBCwUAA4ICAQCT+voIwERe6otfHA1zu6NgdZkQ4F0JYXiffJEP
lOFaUhIK42k7lTm45D52ov+POAAd5ob+4CBr1sAtTO3nvC3ol8p0Inky+NfKN0WD
Cxu4nxlpzzcTl1HjyIXripWivJbIky8aYgcEjfxkqomeo69/7mnFYxSXseq69/BU
yDHAC2G/jaLKe4q0BIo4i3a2+LJ50fz2jrBz7NxzuUslOhqWL3s+f27hnWkD2ces
JW1+rJDqdnobLnBSIXp/8Sits8q/OhVYAVUdjNjNWw4SV6Ig1RAhoJTh3THjWtTa
DePavgy6hJiD+HLluAvy4fSaOxm7stQjMkuZEEalCddB6tXfNw/CtnIBQBSF6cWa
2QBbq5kDhwHYG2+HOcV37/3Q8h28QPDIR8mbKkO5AiBbDz0xkti+0N1ZDn+qEvGv
83Z84+sMf67/EnV+ryyfEhU2KpZcGH6rwYJc0FGjp8mZFdrSEFLp0uZbVf274TXD
3yFjKCQktJgIqt2RMYEtRQNgC3NBPGvDFS7q6S7HFZDI4EqtwGp7j0LfXofViyYJ
wJ/u53KMcbbCCjHhJ/d7bBNV20rw3yTNITrDXX5dhWQ6fNfcOqXlXhpct1AlqCHs
p0TGlNicxb2nnhRuh1U63VkBhl3k17vClDiwqHYwe7fPzlmj6kZMaXnpK6rd4kg7
1anh3Q==
-----END CERTIFICATE-----
`

	invalidCert = `
-----BEGIN CERTIFICATE-----
MIIE1jCCAr6gAwIBAgIBATANBgkqhkiG9w0BAQsFADAaMRgwFgYDVQQKDA9rb25n
X2NsdXN0ZXJpbmcwHhcNMjIwMzA4MDgyNTIzWhcNMjIwMzA5MDgyNTIzWjAaMRgw
FgYDVQQKDA9rb25nX2NsdXN0ZXJpbmcwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAw
ggIKAoICAQDcxjHjleVMebZZkC3EASExq2blPHi0Ie/lODeBfcaldVPKp7iihMIX
VtnbmzRxPnBwnUvy85TUrBDsyDHAfcxVlRzzEEMNbHcOoTzLIMC1KLo9Ksk/Zj5j
o3gGGy3mh2DlYR2gRCShoquBotlvGxXLc7mni2RrY6j0ib2MEdFQCZTOadpmRep1
ZnMhTM195de5NpzBg0em9pzkt8PFlovzgvOmFkRm0PVizvNhwYVKd1WO/gnFx0a9
8KlBZnD0tes72qJresaIzsFGiIuMfmjtL7SmRlFr8gEfMPi9V8n0m4lpUabPvGj/
szkwwcu0KZR13Zjddv2kGiJh0UOTYSf0q2fewyG1IjUo4Zc2R8ezyJGycWd2UFK3
mNUTaat4mZrYgCr1U49kUNf5nOniCpP+G5Z652Kg1ZcLwdjWf+sBf49+1izKgKA8
Utv/VDxqcm4ZbLvk/QLK+j+qA4CYDqm9XSGlFvXAYKIcUdy++rPiT/R9R56XU8Pr
cZmedEvJWJ6B93nSVWwuLtMdQPT4HR5++Onff4+25eHCCDLhXuHauJ7ayRPUEaut
td+x8wxUtLbiGVS8Rxy6pSZFWsaNNEwq/pDcXIWv6GUIERI64FCYhEVYy3rsbqj+
kL+K7tAgIYgeGvU3PMP28aOZzuIePrrJP0qJEL/drjc38hBC9pQ17QIDAQABoycw
JTAOBgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDQYJKoZIhvcN
AQELBQADggIBALQjU9lUszItQYpUcfSwWNLiowP+IFRqsgN3zefzwyIpRqxsbway
FFOObgzCH8bkjMq8rLt6BKn/JAmR/NMuOSY5oCL64wHBTtsfRBzkSlTWC4Pv2kJy
84zXRauwQotS+PB3ZifhdL4iJ3i1wBhiZtZbt3WB1v6q2HCGZNEoDR66QjLCdq4L
90OO5cTbq8MYGNGVYEPu+t80FGIzZ4V176+xAEfi+pBQCNsZr27Mswi59HmCyThi
VEGAmK5DTxUXWFhR5VDEPxiyEO6QAcae9cMVni3znXFvaw9K/VqY+aGj326ThfGu
t2HEtKIVck7fpOA4AWA0zVw+UUFv5FNDy2amluCP1WlYgbvNmrChlGU4WCUp+twv
yKiZ8aquEsxhhIqvz714OGHe5WxxoMejzRl9cMaunhBqSv3jYPPwLptrktbhrjJJ
zdeTAgOI/zC4eTV6s1IKnuHYyXm3FtJ7Sl/tPbDkRTI1dgGPT4yQzbwLPpdPsvB6
bXvl4twKM0ygKuy47SgOnvGC1NVePZEWzEovppy7wL4vtJWm+I4/rmBa9o69xonK
Zm4rDc7UfM+m3t9OKvyBO1iurXSxQOksybojuCGGyPwFHJB8teuzt7UqL5tZMEz3
KfscJPjs+Kb+JEWNMprtJD0Odnp/TPEGfJbHEV8Sw1BID56TXiuGVpeg
-----END CERTIFICATE-----
`

	expiredCert = `
-----BEGIN CERTIFICATE-----
MIIE5DCCAsygAwIBAgIBATANBgkqhkiG9w0BAQsFADAaMRgwFgYDVQQKDA9rb25n
X2NsdXN0ZXJpbmcwHhcNMjIwMzA4MDgyNjUyWhcNMjIwMzA3MDgyNjUzWjAaMRgw
FgYDVQQKDA9rb25nX2NsdXN0ZXJpbmcwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAw
ggIKAoICAQDL57Tig7wZ9EUhsgConAfCBhVT0BspZHYICYjL1MTBmKYb8q3CWgjx
xQQ6rcEZhAVgtcuxyoH+OM8dCkIj2uuNKO2bn9UoHHo3Ofj7YMivL9pfx+tsv0U0
APofePmDMdg9zSF5nUYUrsNeFX2Mzoaj3sEjLSqIR5RRgcinIaL/IlPDK9//W0qn
nkXMF6Oqpcmpk2AvXwDY2EAc5A3gN9nWfT/3dfh7PaisHLc8LoH3xIbmZESm73I5
kboE1a8qgims3ZRrI4q1I6/v8YegjVO1eKElWjHqrFUzG0Z5/FasJm7j5dutGama
c4Eco36FLazBaFFOA9WKQbFOmmccXOxFk1RM6IsNDiYWlRrCg0o9+Jq9wlcTiKN1
TFYLi7lLYarsqcJAHkYRhkvZAA9Pwvetb128PKkRQApLGHoq0HiT8qdnzKRCrW/I
rAtxttMkd6xa/kYu4bciKM1RTKTEm6+Pkrg9EiFHTs2DLjYd/hYJbI22WChDlY3M
GCou7yc47DP38ov13aTl5AiXPgha4olvAZJUwdZdoC4/cHJSuvrb1YYQ/Xgg7dNv
FJCryG6bwszK3R0fVjMANQ076r54IsnQR+dFyOqkG/rY5/WYEWRzfmfiWMwzYQ9S
tDfthQ7lgGV0WkyoRa5KGwxHoRi/ZSfGs8ir7G+OD99dQeGD5puUNwIDAQABozUw
MzAOBgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDAYDVR0TAQH/
BAIwADANBgkqhkiG9w0BAQsFAAOCAgEAaHPbaj4zb0B4rv8sjExJxHcUiI5XRTUa
ZW9tUvvyGR6fhWrP8S4+aGkP7gD4cZNFHBAge++Bsjvo5SANgqQdV9BAUFB3adxv
dr60ddw5UHhvQhl3h42RsFJvgwaqQVJF3eDRsbcKb2paRNbNKxRWP5eyflgOYR5a
DiSWsqA7yGCBntwL5wIwN9nNv1Tt5yDvg+WQzRgzKgFn9gyNruxeYBpUa/HMJAxY
abzylaDpTeJyxbIaZ/LSY0oKgTzfN/gmPB79DD/n8iKSfMYO71opyEsogeozle64
0nHRz9DE72ryt+ZufkyWeN/Rlrh3ohD/JqwLHHWE4nIOB+c7kTwqc7kJHBmA+5Jn
JvOu0hsd2qv6IIV45KQZ4DKYQ+KOedWPO2eNjY+HrzT2He3Mp+6N80UnXWA+gNEu
bkNHNF9PtYDbEIeUFN4GyMaMVtYMvlJkevW43hH6GyLlSVfTp9nVXVxpyJAmH7t/
TqRwBYiU3blVS4pCYE6UJhlFSWjBY5vm+54ShL2o/aKP+HF8Dme1dznRYdyC5AAt
PUWXk+7W37+68xusiRippoH6N4aky191S0dST8/XlTbIIPtq9ImFngrL0JGOAYnY
4byupLnu1JgInM64Ca5Ps04arQGTiMJ7I3Vk6NIbyuA254i4J/ZH6FHwDuIsSmzv
Vd2k4e2sXh4=
-----END CERTIFICATE-----
`
)

func TestNewCACertificate(t *testing.T) {
	r := NewCACertificate()
	require.NotNil(t, r)
	require.NotNil(t, r.CACertificate)
}

func TestCACertificate_Type(t *testing.T) {
	require.Equal(t, TypeCACertificate, NewCACertificate().Type())
}

func TestCACertificate_ProcessDefaults(t *testing.T) {
	cert := NewCACertificate()
	require.Nil(t, cert.ProcessDefaults())
	require.NotPanics(t, func() {
		uuid.MustParse(cert.ID())
	})
}

func TestCACertificate_Validate(t *testing.T) {
	tests := []struct {
		name          string
		CACertificate func() CACertificate
		wantErr       bool
		Errs          []*model.ErrorDetail
	}{
		{
			name: "empty certificate throws an error",
			CACertificate: func() CACertificate {
				return NewCACertificate()
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'id', 'cert'",
					},
				},
			},
		},
		{
			name: "invalid certificate fails",
			CACertificate: func() CACertificate {
				cert := NewCACertificate()
				_ = cert.ProcessDefaults()
				cert.CACertificate.Cert = "a"
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:     model.ErrorType_ERROR_TYPE_FIELD,
					Field:    "cert",
					Messages: []string{"'a' is not valid 'pem-encoded-cert'"},
				},
			},
		},
		{
			name: "valid certificate, but invalid CA fails",
			CACertificate: func() CACertificate {
				cert := NewCACertificate()
				_ = cert.ProcessDefaults()
				cert.CACertificate.Cert = invalidCert
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "cert",
					Messages: []string{`certificate does not appear to be a CA because` +
						`it is missing the "CA" basic constraint`},
				},
			},
		},
		{
			name: "valid but expired certificate fails",
			CACertificate: func() CACertificate {
				cert := NewCACertificate()
				_ = cert.ProcessDefaults()
				cert.CACertificate.Cert = expiredCert
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:     model.ErrorType_ERROR_TYPE_FIELD,
					Field:    "cert",
					Messages: []string{`certificate expired, "Not After" time is in the past`},
				},
			},
		},
		{
			name: "valid but multiple certificates fails",
			CACertificate: func() CACertificate {
				cert := NewCACertificate()
				_ = cert.ProcessDefaults()
				cert.CACertificate.Cert = goodCertOne + goodCertTwo
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:     model.ErrorType_ERROR_TYPE_FIELD,
					Field:    "cert",
					Messages: []string{"only one certificate must be present"},
				},
			},
		},
		{
			name: "valid certificate",
			CACertificate: func() CACertificate {
				cert := NewCACertificate()
				_ = cert.ProcessDefaults()
				cert.CACertificate.Cert = goodCertOne
				return cert
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.CACertificate().Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.Errs != nil {
				verr, ok := err.(validation.Error)
				require.True(t, ok)
				require.ElementsMatch(t, tt.Errs, verr.Errs)
			}
		})
	}
}
