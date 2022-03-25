package admin

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/stretchr/testify/require"
)

var (
	goodCACertOne = `
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
	digestOne = "a239094c44503b6a75071a098d6ef2fdbf1009343f60bbdbb17f52701cd823b1"

	goodCACertTwo = `
-----BEGIN CERTIFICATE-----
MIICtjCCAZ6gAwIBAgIJAM5Fd3HKorRaMA0GCSqGSIb3DQEBCwUAMBAxDjAMBgNV
BAMMBUhlbGxvMB4XDTIyMDMwMzA5NDkzOFoXDTIyMDQwMjA5NDkzOFowEDEOMAwG
A1UEAwwFSGVsbG8wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDZxtUd
K3kn6JFRCznlbacn9qc0/vYUfJM9CGH3nlLoonI/6Bc+4ZRGKXPW0RVFynWq0cdq
kpgxX9dcnKLXLI2DHhzY1qykW/ieLSjG5twyU0d3SxRr4gYy29KXjW7sWUEUrMmn
zW+ody2YX3d+ppkUJJsmqucH4S+hUeIC8gm6cLE7H/R2GKvy/Ny0vsCT1xVXCFh5
ko9FT1AhSierbPx2Njr7dNaKyuqR2gmhC7Aq6xLwDVBeMPoUXsLM8eXBntIrrQLw
5FkHJIidLav5UgCYFKal+9bbtYdK2tO5xA+tu7BsKC1u0zZtEcX2/CzNPNJRitAR
DIZMipzHeUrjtCfbAgMBAAGjEzARMA8GA1UdEwQIMAYBAf8CAQAwDQYJKoZIhvcN
AQELBQADggEBAG90erz5GhzpahTm9KC5hKkyZbyUxvKLMkQSSB2S7go6irA9TAIW
DOmRaQoVHh+BxX4bwqdzKhy24IPg4oe3xvD8RXJAtlersdZt3Hypfcfw9y2E3Bsr
CmwpdYRh7iO++FqYXqMYWsHjvaiXmyXyO+ZLCDH3QUFKJEHC7ivP9zTTjgDqvLxd
52qJFHgU99zcQPYJy4r6OhrkYTyCsmhY5+lmpxz6AENqGk2mQCwike2LlRJRy8fB
RZzSYCLlEcxMDULSa0e1i6cQBbARUdDAaGLGZoNxCXB9i+6L9l8XsZgoCJXdjI8O
XKfF8+VaoZcdgcddAc/pG1AQdEZVMEmSRcA=
-----END CERTIFICATE-----
`
	digestTwo = "bbf6b98dcb52a3fb38e915fe3f6ce7b06ce32f83d86e7fe9b1ec6d3c78b41a95"

	goodCACertThree = `
-----BEGIN CERTIFICATE-----
MIIFEDCCAvigAwIBAgIBATANBgkqhkiG9w0BAQsFADAaMRgwFgYDVQQKDA9rb25n
X2NsdXN0ZXJpbmcwHhcNMjIwMzA4MTIxNjA3WhcNMzIwMjI0MTIxNjA3WjAaMRgw
FgYDVQQKDA9rb25nX2NsdXN0ZXJpbmcwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAw
ggIKAoICAQClyPtkuXCD62IpqT9QMZWObrDkKHL0Aln5qeXmaVxItp/J6H/Ksob5
tAM6cg0ow682sDT7x45fnKKfxOge+BdPd/K2eMFm2qEgTWGWVJE13Di/CibDhwA3
Y5+UJJLyjoZKOpy7HWzBS1rc0AipeyKkaDe6dXNkT9gysUE1A9L4PiFWgXk2D4nJ
D3cKduXU91xptL/6Ww89X+9a1rOYPSayl9AhOSF6UDc5jPpCCoilS3xM2PeyzOuQ
mHMxq1IPGwInpKgQvPJ4nWosIsZwfD7z2RasDgOk+Qn6WphYNXHc6Vv1kAA6pggT
gUHBYQCymrBnBBzNsK9GI0YPFhm9c9cErtgcaNkt43xU7Wbxg4rhIDP5jMH9sTew
8JGfCqBwjD/DS87seodd/pL1LmA97eK0IZUm4IbsGl/Ms0aW2MQvIWQTm7mX913w
OleJpQtktOAoOLXuWX7TvJSyER/2Ea+rmk2G4Gqpe67azIdQxbMy0j5Iol7IBesz
GJZWVGcwFDFzJzsvrpiSK4Alb7GFRNwlM+h1Py9qFuKqgutwb9GvJ83f08ipAp1s
ipJxBvVe7XqmWq2oQjFw5kSQYL0vUrXCLS7oDjnjStaPdOXqvvNHGhkTX3aLDBar
ybLcLMaefDom7q7gTkB7+yxnMnfLfW8kdESWCce7e218GenP88SYqQIDAQABo2Ew
XzAOBgNVHQ8BAf8EBAMCAoQwHQYDVR0lBBYwFAYIKwYBBQUHAwIGCCsGAQUFBwMB
MA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFHZwAuaBCTk4uJiosDp6x46qvnxG
MA0GCSqGSIb3DQEBCwUAA4ICAQB6eP/yfXvSMpv2Q/SNCI0osj3nvhRpHcksO1dq
5hEbSs//WiEtIOPdFSEIV1rIU6fbEyVHF3s+77p4WFtAjBhOFa5HIC2Hdx9v4QMf
mZg4ycB2towUsQmObOi5KHZvzEVmRoFbsFE3Tg/0z9XwM6NsAjHxmy2eDvaMep1w
qwJXNbsw1qGP1brMs7GOjwNMZ5EnzOEFWVFQIcrZ5aFRI6FijPgwc7N26i293aK4
nSthZ0DgrKyCUBmDOJMn+U6U/FCOCPp8dfnvrwTeSubM9OGlij5aTK4G3icXfzqJ
1U6qvZV/3j+DolZlke0HHnoWEdq/BB+qvlNiVcbeGjlMH9t9EfsKTLsYBYuOOW/0
JwBXAlP21Nsjc67DLGnCcQIxgwIyrDyA2b6JT1qxJ6ptro0LBp1/iJxDHuqiY8ZC
vwK+xhGyEfIlEbBY1F5LZI1By8Mhr42+KrIbHwN0sfNd0ayBIEyV8ZdUxyN+/P37
Qdnv7qOY6ahKHBWDOYbMol6cqDYFGrzZ3M4HaiTxeMivlmhTfIWMdGdiIPtpw7zw
MV8KnKj9P2aKSiIaHoHL5oG3CdaevecpXptyU0w9sWQ0CQLSLNYqgXMPbrfGFWD5
eemwDiB6PVaSnWFtRqiYVTzEflk6Not8/XguC9hN80Lb3yEj63pgrvyl34Nu0GaQ
4TJciQ==
-----END CERTIFICATE-----
`
	goodCACertFour = `
-----BEGIN CERTIFICATE-----
MIIFEDCCAvigAwIBAgIBATANBgkqhkiG9w0BAQsFADAaMRgwFgYDVQQKDA9rb25n
X2NsdXN0ZXJpbmcwHhcNMjIwMzA4MTIxNjMwWhcNMzIwMjI0MTIxNjMwWjAaMRgw
FgYDVQQKDA9rb25nX2NsdXN0ZXJpbmcwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAw
ggIKAoICAQDUhKBe8Dv7NM5UcQgTh+YPTPYEIeVJNEcDwMCwoRjaxcxJF3RTUguC
X5Rb66oo7eJeXqpgPWmTyGCFNbS8oSolyc6mk1c6DcT0gpiv5tJmYGPLnY6lbPic
iZt/Q6wfXhHm7CTT9Kld//SaIsPk6QCXHHOdhaTN/PIgFPzrcjkb4ZdoyBcdHdLl
0SEAnHbyJOlKQHC6PUdPGDuQxytQPAw8HFpDbqurCXdxyuSvp3kiiZrO+bcY/Bk7
x3NKKXjJloPq+mm6960CPCDWDz8xtwi4TIbpo20WokJq0Fp7v79ezvysdSXHvpc3
Dw5SU0RKnh3i/Id71ldEV72FVUWVggnkRMpshtyXAuJVng9vdBqZ/vSIBnMG1m26
+W0Aqb/w2W/poBSdwTnj0aNvbZMiS3jZnivJI+EY5WT648tZBfr70OiUfGXd/6G9
mCVM/ApHC1WfdZZL3LfBlz5q5lS3A84LYhdNIjpJ2+A8Uml35ui7ynZxC8n6/WD/
6iVnqLenF3MxXfFhykgrCFlR9vgB1cHMzP4QQR7izsuSczawVcxy5tmryAG4PrcL
0SJb6EtZWTvhlO9t5RGJAkXnxzCj/hC2znswSD3GehtYOmuWoJ/B2vf0Hw5K2WCD
ocMI1/pqoGPbSG6fx+eIVTXBpqhq5pyoKRYbxHSeo3OXyD1sY8zMJQIDAQABo2Ew
XzAOBgNVHQ8BAf8EBAMCAoQwHQYDVR0lBBYwFAYIKwYBBQUHAwIGCCsGAQUFBwMB
MA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFMhPxlkgaVzSZEEZ0UUYWVXkNO2H
MA0GCSqGSIb3DQEBCwUAA4ICAQBlYl3GWOVgjQ1bhymOtX76Zh/FF5cmD6hzDzi8
hZafVObPIovpEcLeL6t0qdXosSCxSbgeO+7Mh8Cc3djFiM23JINjNuDTTOhqf0Fq
je1hxy9uagI9wULlN7AEI/xpcFCkUWLL833VztwItJRJFpZ37Fd8lQlBwRPkA6RN
M0yxUmGLAG5o6wNIrvQkK5BqMhptT6RSzQyTpa2JmYLOTLjd8uzuG+zjxrKMxClo
dZe6tsU6dGQow/wtlwjPpXoz5AU6FKd+JVBcEXdEpCCWMUTzDyNxemGHXoFGoZ97
zJcR95yGL5jGAl48YDv+lIHgRUvY0sPLGl7pspGeCynC0/qRkH7mSU/EfN8QKrhU
H9CRJQefdsR8kMyNn4ThqR4pKbSlOGgyhsHNNNv2MJbWSffbCXnmZh1ys9pwMYsM
Hq1oEhSQbn9ylMJaNurMrd/B3tlwjjNFldwADYJ0HMUInakmHHWtm+/MGpfJdWAL
37PCIXjij/lZyY0NlqMdhRE++NRemPb2eEMqXfXq8EMMvgIDZJGEgoggNhgTPbql
/+XbaGRvFnSUCm/1OgxtIomLm6i+XmB71/gWLAL5OzwpYU7cnbpNf9nLJ5FVasDg
qF1M2HH9FXi61YiZodg1mxUuCcXIGUMv0erNUtMLtyuKmbAsC/sLAUUqin9y3np+
mtvVeQ==
-----END CERTIFICATE-----
`

	badCertOne = `
-----BEGIN CERTIFICATE-----
MIIE4DCCAsigAwIBAgIBATANBgkqhkiG9w0BAQsFADAaMRgwFgYDVQQKDA9rb25n
X2NsdXN0ZXJpbmcwHhcNMjIwMzA4MTIxNzI0WhcNMzIwMjI0MTIxNzI0WjAaMRgw
FgYDVQQKDA9rb25nX2NsdXN0ZXJpbmcwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAw
ggIKAoICAQC9Fn00WPHFrKAMg+eLo6OzoMNynXMyUqijoK9Fp29mGC9b2IjpRazK
EOkJ0BUbzkLZGuLKDq0OqQYO/EUROK0lkByn0DP8wI0GqpaUkDlaT+ISRQhlrafY
dpRo0bTceHKcID1o5Frc1JQ2kdiLitAWUt1APEbm6Mo+N8f4Re3vlQNNonLHlDuo
oxH1fdYXpOSfmrZMjhO1Ljy00pEC/He2qB+gZpUwPhBks4p+DLd8EQYD/KWzVXg9
mQSCJllbgyAEqkWqiuTvu19PMDjEMMEQVewvz/Rf1DeUXBPXmdCyG3q4/8aHgRA7
rClD2188QRVtBTPkU0nUFfLKWe4IRW3X1IR9xruXuuoLNTp05b1ny9frZACiE9Of
A5fDQPooUANF5cRsTck88rhbTDz0DzGtx2S1mP+hg+2knyoCHz38k3Zo5TB1wjLO
wt1er6vdsARttjzjVZyNPvAMncD0zTNu4VVUHKxzpNtzkXWEPVps5ozuuCbgoqvg
wy5RkjlQSmHqEaI9F4tm2KrQJNIpAvN7CFd1QAH/CxK4i1dsB7dSaFBsctvsjqD3
9125BPfT305FxkR6U0U00gq0tbbl2cnrGtcXD+BhXCln276SfRMJ979rzTMy5EBs
dDfT2u5mIjlQgqYayQvxSY8Drbrt+dPLt7gTf/YdjHdrHTdtWdmCYwIDAQABozEw
LzAOBgNVHQ8BAf8EBAMCAoQwHQYDVR0lBBYwFAYIKwYBBQUHAwIGCCsGAQUFBwMB
MA0GCSqGSIb3DQEBCwUAA4ICAQCtKeZ3DvAVabKNb4uEL2Y7fMURdp/XAFnWxcRl
3+HFTgGYHRF2BO7d/3C0dSmCGf8Da3OX4X9GDamE2O2X98gcj3TbZ5C+774R4eAO
9iYCN0ZMlMAUB7EWe+kvStdN5A24V/KSI9N0R6LjWaDvOl+xOlqtDTguj6s11m+D
mJMuAXAPBudBGT1lr6XJWu7yK8Nccb6iqBDe5T4pSAkuR/qF5hNGqdCjyncxViW2
sQsPzMoZuDureLEW9vZ2ZaHGLCsO6fWI2lZyxMrDHpHKLIY8b3bz4DA0tRpj1UQP
O6sdUZuWr1K9s6GASFxkaXyf4VDjk86Je0qC2a4GPFKD9Wx78Pn1KRd8ViI5O2lJ
fKNg+02eDjJL9JrEGCGbVMO8vCUvIlCdv1YKIMHVEYKzv9Qop392dkAOhhRg8Hwj
vgn2yyibZTehEQ32rU9VsM91Z2jnwQ+FbpPVWTN9NdPyyPysDrcvyG/xDAiaKAeD
VYm39P9EzzqJ4+M+zXZmCePu2bzYOAv8lz+SKoMSswBQTUsKPeXuGrJygX1gEFFT
aWdvIEokymPRgwuCy2FWFC45pzixbMLQv96noMy7meAJxffTIER4NVxJEeJv4qnd
jZQMR3Q6NGrbKbO8u+2XwJZGroow/K8KpiI5mFpI7xc8wKvUf0cPbDdr4X8S7mDY
9tPrfw==
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

func TestCACertificateCreate(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	t.Run("creates a valid certificate", func(t *testing.T) {
		cert := &v1.CACertificate{
			Cert: goodCACertOne,
		}
		res := c.POST("/v1/ca-certificates").WithJSON(cert).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.Value("cert").String().Equal(cert.Cert)
	})
	t.Run("creates a valid certificate and check digest", func(t *testing.T) {
		cert := &v1.CACertificate{
			Cert: goodCACertTwo,
		}
		res := c.POST("/v1/ca-certificates").WithJSON(cert).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.Value("cert").String().Equal(cert.Cert)
		body.Value("cert_digest").String().Equal(digestTwo)
	})
	t.Run("recreates same certificate fails", func(t *testing.T) {
		res := c.POST("/v1/ca-certificates").WithJSON(&v1.CACertificate{
			Cert: goodCACertOne,
		}).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "data constraint error")
		body.Value("details").Array().Length().Equal(1)
		errRes := body.Value("details").Array().Element(0)
		errRes.Object().ValueEqual("type",
			v1.ErrorType_ERROR_TYPE_REFERENCE.String())
		errRes.Object().ValueEqual("field", "cert_digest")
		errRes.Object().ValueEqual("messages", []string{
			"cert_digest (type: unique) constraint failed for " +
				"value 'a239094c44503b6a75071a098d6ef2fdbf1009343f60bbdbb17f52701cd823b1': ",
		})
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
		cert := &v1.CACertificate{
			Cert: badCertOne,
		}
		res := c.POST("/v1/ca-certificates").WithJSON(cert).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		errRes := body.Value("details").Array().Element(0)
		errRes.Object().ValueEqual("type",
			v1.ErrorType_ERROR_TYPE_FIELD.String())
		errRes.Object().ValueEqual("field", "cert")
		errRes.Object().ValueEqual("messages", []string{
			`certificate does not appear to be a CA because` +
				`it is missing the "CA" basic constraint`,
		})
	})
	t.Run("creating an expired CA certificate fails", func(t *testing.T) {
		cert := &v1.CACertificate{
			Cert: expiredCert,
		}
		res := c.POST("/v1/ca-certificates").WithJSON(cert).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		errRes := body.Value("details").Array().Element(0)
		errRes.Object().ValueEqual("type",
			v1.ErrorType_ERROR_TYPE_FIELD.String())
		errRes.Object().ValueEqual("field", "cert")
		errRes.Object().ValueEqual("messages", []string{
			`certificate expired, "Not After" time is in the past`,
		})
	})
	t.Run("creating multiple CA certificates fails", func(t *testing.T) {
		cert := goodCACertOne + goodCACertTwo
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
			"only one certificate must be present",
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
			Cert: goodCACertOne,
		}
		res := c.PUT("/v1/ca-certificates/{id}", id).WithJSON(resource).Expect()
		res.Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.Value("id").String().Equal(id)
		body.Value("cert").String().Equal(goodCACertOne)
		body.Value("cert_digest").String().Equal(digestOne)
		resource.Id = body.Value("id").String().Raw()
	})
	t.Run("upserts a valid certificate ignore digest", func(t *testing.T) {
		id := uuid.NewString()
		resource := &v1.CACertificate{
			Cert:       goodCACertTwo,
			CertDigest: "a",
		}
		res := c.PUT("/v1/ca-certificates/{id}", id).WithJSON(resource).Expect()
		res.Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.Value("id").String().Equal(id)
		body.Value("cert").String().Equal(goodCACertTwo)
		body.Value("cert_digest").String().Equal(digestTwo)
		resource.Id = body.Value("id").String().Raw()
	})
	t.Run("upsert an existing certificate succeeds", func(t *testing.T) {
		cert := &v1.Certificate{
			Cert: goodCACertThree,
		}
		res := c.POST("/v1/ca-certificates").WithJSON(cert).Expect()
		res.Status(201)
		body := res.JSON().Path("$.item").Object()
		id := body.Value("id").String().Raw()

		certUpdate := &v1.CACertificate{
			Id:   id,
			Cert: goodCACertFour,
		}
		res = c.PUT("/v1/ca-certificates/{id}", id).WithJSON(certUpdate).Expect()
		res.Status(http.StatusOK)
		body = res.JSON().Path("$.item").Object()
		body.Value("id").String().Equal(id)
		body.Value("cert").String().Equal(certUpdate.Cert)
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
		Cert: goodCACertOne,
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
		body.Value("cert").Equal(goodCACertOne)
		body.Value("cert_digest").Equal(digestOne)
	})
}

func TestCACertificateDelete(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	cert := &v1.CACertificate{
		Cert: goodCACertOne,
	}
	res := c.POST("/v1/ca-certificates").WithJSON(cert).Expect().Status(http.StatusCreated)
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
	certs := []*v1.CACertificate{
		{
			Cert: goodCACertOne,
		},
		{
			Cert: goodCACertTwo,
		},
		{
			Cert: goodCACertThree,
		},
		{
			Cert: goodCACertFour,
		},
	}
	for _, certificate := range certs {
		res := c.POST("/v1/ca-certificates").WithJSON(certificate).
			Expect().Status(http.StatusCreated)
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
