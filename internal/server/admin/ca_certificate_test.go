package admin

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/stretchr/testify/require"
)

// NOTE: certs expire Mar 26 19:47:33 2032 GMT.
var (
	goodCACertOne = `
-----BEGIN CERTIFICATE-----
MIIDkzCCAnugAwIBAgIUYGc07pbHSjOBPreXh7OcNT2+sD4wDQYJKoZIhvcNAQEL
BQAwWTELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAkNBMRUwEwYDVQQKDAxZb2xvNDIs
IEluYy4xJjAkBgNVBAMMHVlvbG80MiBzZWxmLXNpZ25lZCB0ZXN0aW5nIENBMB4X
DTIyMDMyOTE5NDczM1oXDTMyMDMyNjE5NDczM1owWTELMAkGA1UEBhMCVVMxCzAJ
BgNVBAgMAkNBMRUwEwYDVQQKDAxZb2xvNDIsIEluYy4xJjAkBgNVBAMMHVlvbG80
MiBzZWxmLXNpZ25lZCB0ZXN0aW5nIENBMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEAvnhTgdJALnuLKDA0ZUZRVMqcaaC+qvfJkiEFGYwX2ZJiFtzU65F/
sB2L0ToFqY4tmMVlOmiSZFnRLDZecmQDbbNwc3wtNikmxIOzx4qR4kbRP8DDdyIf
gaNmGCuaXTM5+FYy2iNBn6CeibIjqdErQlAbFLwQs5t3mLsjii2U4cyvfRtO+0RV
HdJ6Np5LsVziN0c5gVIesIrrbxLcOjtXDzwd/w/j5NXqL/OwD5EBH2vqd3QKKX4t
s83BLl2EsbUse47VAImavrwDhmV6S/p/NuJHqjJ6dIbXLYxNS7g26ijcrXxvNhiu
YoZTykSgdI3BXMNAm1ahP/BtJPZpU7CVdQIDAQABo1MwUTAdBgNVHQ4EFgQUe1WZ
fMfZQ9QIJIttwTmcrnl40ccwHwYDVR0jBBgwFoAUe1WZfMfZQ9QIJIttwTmcrnl4
0ccwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAs4Z8VYbvEs93
haTHdbbaKk0V6xAL/Q8I8GitK9E8cgf8C5rwwn+wU/Gf39dtMUlnW8uxyzRPx53u
CAAcJAWkabT+xwrlrqjO68H3MgIAwgWA5yZC+qW7ECA8xYEK6DzEHIaOpagJdKcL
IaZr/qTJlEQClvwDs4x/BpHRB5XbmJs86GqEB7XWAm+T2L8DluHAXvek+welF4Xo
fQtLlNS/vqTDqPxkSbJhFv1L7/4gdwfAz51wH/iL7AG/ubFEtoGZPK9YCJ40yTWz
8XrUoqUC+2WIZdtmo6dFFJcLfQg4ARJZjaK6lmxJun3iRMZjKJdQKm/NEKz4y9kA
u8S6yNlu2Q==
-----END CERTIFICATE-----
`
	digestOne = "34e0f1f3d83faefcc8514b6295bc822eab1110dc120140ddf342c017baee8c0f"

	goodCACertTwo = `
-----BEGIN CERTIFICATE-----
MIIDkzCCAnugAwIBAgIUGAsxwC6yfwavG9MX3EDvI6jbqpUwDQYJKoZIhvcNAQEL
BQAwWTELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAkNBMRUwEwYDVQQKDAxZb2xvNDIs
IEluYy4xJjAkBgNVBAMMHVlvbG80MiBzZWxmLXNpZ25lZCB0ZXN0aW5nIENBMB4X
DTIyMDMyOTE5NDczM1oXDTMyMDMyNjE5NDczM1owWTELMAkGA1UEBhMCVVMxCzAJ
BgNVBAgMAkNBMRUwEwYDVQQKDAxZb2xvNDIsIEluYy4xJjAkBgNVBAMMHVlvbG80
MiBzZWxmLXNpZ25lZCB0ZXN0aW5nIENBMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEAvnhTgdJALnuLKDA0ZUZRVMqcaaC+qvfJkiEFGYwX2ZJiFtzU65F/
sB2L0ToFqY4tmMVlOmiSZFnRLDZecmQDbbNwc3wtNikmxIOzx4qR4kbRP8DDdyIf
gaNmGCuaXTM5+FYy2iNBn6CeibIjqdErQlAbFLwQs5t3mLsjii2U4cyvfRtO+0RV
HdJ6Np5LsVziN0c5gVIesIrrbxLcOjtXDzwd/w/j5NXqL/OwD5EBH2vqd3QKKX4t
s83BLl2EsbUse47VAImavrwDhmV6S/p/NuJHqjJ6dIbXLYxNS7g26ijcrXxvNhiu
YoZTykSgdI3BXMNAm1ahP/BtJPZpU7CVdQIDAQABo1MwUTAdBgNVHQ4EFgQUe1WZ
fMfZQ9QIJIttwTmcrnl40ccwHwYDVR0jBBgwFoAUe1WZfMfZQ9QIJIttwTmcrnl4
0ccwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAqhJ+aEQT8eSy
bzvcXqz6Ou8EQk3oWRXuLb6fNRGZkFE2b2xysxyOaXoK04XddE1Wx1WkaVT3poC/
t09EAnMluxjpSOkQ2nwsSXYtorqJGYDO4WK2QP+mwnUD2VfDvi9p1qZB1vS+9aM0
LtP3Qp97CKpNajSHc8edDtEG55vMf2j/BVvEeAhKPqKF/oRTE/hRt4N3u5tfE219
80+M+JaQF0IGJMId1/FxTtiUI+uBpumRveZF7AI4aLWTmKiEV65rmM/Z47m+hBJr
4y6QtTJtCXGPXMkkX87KivIRthulPqY7OJH8+2Kz1v64a2XwbKdJ6mq4bGX02DDD
8Z+8uIMIzw==
-----END CERTIFICATE-----
`
	digestTwo = "e30fce928184a3991553ee72192907e9c64c377949a0cb1a990751bbce5919fc"

	goodCACertThree = `
-----BEGIN CERTIFICATE-----
MIIDkzCCAnugAwIBAgIUHkJ9/3l6zONzta31OEmG0e043ckwDQYJKoZIhvcNAQEL
BQAwWTELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAkNBMRUwEwYDVQQKDAxZb2xvNDIs
IEluYy4xJjAkBgNVBAMMHVlvbG80MiBzZWxmLXNpZ25lZCB0ZXN0aW5nIENBMB4X
DTIyMDMyOTE5NDczM1oXDTMyMDMyNjE5NDczM1owWTELMAkGA1UEBhMCVVMxCzAJ
BgNVBAgMAkNBMRUwEwYDVQQKDAxZb2xvNDIsIEluYy4xJjAkBgNVBAMMHVlvbG80
MiBzZWxmLXNpZ25lZCB0ZXN0aW5nIENBMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEAvnhTgdJALnuLKDA0ZUZRVMqcaaC+qvfJkiEFGYwX2ZJiFtzU65F/
sB2L0ToFqY4tmMVlOmiSZFnRLDZecmQDbbNwc3wtNikmxIOzx4qR4kbRP8DDdyIf
gaNmGCuaXTM5+FYy2iNBn6CeibIjqdErQlAbFLwQs5t3mLsjii2U4cyvfRtO+0RV
HdJ6Np5LsVziN0c5gVIesIrrbxLcOjtXDzwd/w/j5NXqL/OwD5EBH2vqd3QKKX4t
s83BLl2EsbUse47VAImavrwDhmV6S/p/NuJHqjJ6dIbXLYxNS7g26ijcrXxvNhiu
YoZTykSgdI3BXMNAm1ahP/BtJPZpU7CVdQIDAQABo1MwUTAdBgNVHQ4EFgQUe1WZ
fMfZQ9QIJIttwTmcrnl40ccwHwYDVR0jBBgwFoAUe1WZfMfZQ9QIJIttwTmcrnl4
0ccwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEALvETXbqw9sM9
HKxbH442G/1S44Qp6+cerooShApR07rcTz8thU8hY/r/tQsX9MTPPkX+mnjAQwtO
zqoaQOWkRopuIBwq0/4Inke8JGuTFnRjByLkT/2KLKLluiyO5Dk45OnTzOb463Ln
Sxd9UQJhbMYSw63aszR8fustbp8JbjcHfh2O8kfwjfi8elRnTHEKpUUTZ1917IfK
PXE2n5MM0Iup6Hso9YaxjvBBSvIHlvkfcSAxJNhZwPgy0On+i29DcVDPksRjl6D3
0nhUF1li16pqx3quThoy75oOX+EOofuM/uyQZLIzyKDXssozVntztz355pd5fPDH
ku29Tc6GOQ==
-----END CERTIFICATE-----
`
	goodCACertFour = `
-----BEGIN CERTIFICATE-----
MIIDkzCCAnugAwIBAgIUX5ImpR4vO12sjOVbuMm7FV47NqUwDQYJKoZIhvcNAQEL
BQAwWTELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAkNBMRUwEwYDVQQKDAxZb2xvNDIs
IEluYy4xJjAkBgNVBAMMHVlvbG80MiBzZWxmLXNpZ25lZCB0ZXN0aW5nIENBMB4X
DTIyMDMyOTE5NDczM1oXDTMyMDMyNjE5NDczM1owWTELMAkGA1UEBhMCVVMxCzAJ
BgNVBAgMAkNBMRUwEwYDVQQKDAxZb2xvNDIsIEluYy4xJjAkBgNVBAMMHVlvbG80
MiBzZWxmLXNpZ25lZCB0ZXN0aW5nIENBMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEAvnhTgdJALnuLKDA0ZUZRVMqcaaC+qvfJkiEFGYwX2ZJiFtzU65F/
sB2L0ToFqY4tmMVlOmiSZFnRLDZecmQDbbNwc3wtNikmxIOzx4qR4kbRP8DDdyIf
gaNmGCuaXTM5+FYy2iNBn6CeibIjqdErQlAbFLwQs5t3mLsjii2U4cyvfRtO+0RV
HdJ6Np5LsVziN0c5gVIesIrrbxLcOjtXDzwd/w/j5NXqL/OwD5EBH2vqd3QKKX4t
s83BLl2EsbUse47VAImavrwDhmV6S/p/NuJHqjJ6dIbXLYxNS7g26ijcrXxvNhiu
YoZTykSgdI3BXMNAm1ahP/BtJPZpU7CVdQIDAQABo1MwUTAdBgNVHQ4EFgQUe1WZ
fMfZQ9QIJIttwTmcrnl40ccwHwYDVR0jBBgwFoAUe1WZfMfZQ9QIJIttwTmcrnl4
0ccwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAUARk+ZYtQf9z
GzOO4pU5+PKyTJ/QZL3Ul1xqm8yCBbdLc1b9WSt5LYzW0AGDWLr5G299hbHyy+E/
wJXDCIQKTbE4HszrueZXxJ3dgWExHthExTVwNkg9xc8hXOiDtc0TR0IigVyhBJRt
zik8vgo7yFSOEUhgmgco/b+hh8TZ3bVJ/lxDTZu98+1hmCX/T0rUGGBnhSd/wR1J
pRujTJZONfA/x9a0ZB7lYuVtiB27il/9SGGSzk5I50UFPtNGoc5cEYru7mqjHMmz
IOZKPJuDZXcunEf/ZCqUqXsp9dXmgtVTFB3JKTODFSbMcohEfVgn7a/mJcCseCwA
iDbjPTAXNw==
-----END CERTIFICATE-----
`

	goodCACertFive = `-----BEGIN CERTIFICATE-----
MIIDZzCCAk+gAwIBAgIUSXskUBcsOh4tgvhi5MliI8wbpuYwDQYJKoZIhvcNAQEL
BQAwFjEUMBIGA1UEAwwLZXhhbXBsZS5jb20wHhcNMjIxMjE0MTMxMzQ5WhcNMzIx
MjExMTMxMzQ5WjAWMRQwEgYDVQQDDAtleGFtcGxlLmNvbTCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBAMRLfEqflv0IsO7G6N0JPx71vngiRV3khcSBq4Eq
ySUk2r1ojm/mns4lHYSA4YlVeBVk5YIERsL1N7MZWQs0aCODMx1bjINp0oNd5MyN
fbQNlqhLh5N05mI1WV9Hn5KQGXZxxT5AeL1Y56RXry72t6LC5XJuWzPtOqnw/I3D
wgwPcn8WTMOOvqQXkGNnHmDAd+fZyp1MuRaVdl63ELD0YjnQ0tu7TDzjfHff+8nI
KsK505R/EUGRa2yCEwvU2q//ohNBSDKJr4LxyuFLCOtZ1wlTI0skeCoLEN3xjAER
d3q6qIFJW8GfZl7IrG+jLXnEnw6tdzan1YSV9gQXRvB1Ej8CAwEAAaOBrDCBqTAd
BgNVHQ4EFgQUWior1ZGNfrOndXtUIxHLiE4YWaYwHwYDVR0jBBgwFoAUWior1ZGN
frOndXtUIxHLiE4YWaYwDwYDVR0TAQH/BAUwAwEB/zAnBgNVHREEIDAeggtleGFt
cGxlLmNvbYIPd3d3LmV4YW1wbGUuY29tMA4GA1UdDwEB/wQEAwIFoDAdBgNVHSUE
FjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwDQYJKoZIhvcNAQELBQADggEBAFzJ1IgW
qW3W7acduv7XmIzByIZx92Yj/fmjpHtb2kVRLBBkuLm5uD0Ng1jm0RQbZe9K1gdD
zNyG0MSD4qgBFKS3yMHsa2H9svJ6QAAe0dPMX3tSA+UzaJP3yudjCmgxeKYXhzjQ
7HdAp/yjo8mxAu8R6Pu1VvVU/ZTcsFuJLBKWSDdoB3hcdg9YDeS9LUR9Sgii3BnF
MLDBebiwb5+dd0IQpQzMSDc1WNXNSLSHo40WS/mGUvDSLeIKW0V/yWX3hAakyg4r
waHpVltXO7M8pST7SQekx4LBgpm9tDdrH2VSLmpqWDefRoHBFJCcKbhC532NfX0W
FeNAwJWh/ME3d1k=
-----END CERTIFICATE-----`

	digestFive = "d7cae555070f9710f3eca60e64ec4d7c1defe2acb76fae3aede95e2d61fc0dc6"

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
	c := httpexpect.Default(t, s.URL)

	t.Run("creating a certificate with subject fails", func(t *testing.T) {
		cert := &v1.CACertificate{
			Cert: goodCACertOne,
			Metadata: &v1.CertificateMetadata{
				Subject: "XXX",
			},
		}
		res := c.POST("/v1/ca-certificates").WithJSON(cert).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0).Object()
		err.ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.String())
		err.Value("messages").Array().Length().Equal(1)
		err.Value("messages").Array().Element(0).String().
			Equal("additionalProperties 'metadata' not allowed")
	})
	t.Run("creating a certificate with issuer fails", func(t *testing.T) {
		cert := &v1.CACertificate{
			Cert: goodCACertOne,
			Metadata: &v1.CertificateMetadata{
				Issuer: "XXX",
			},
		}
		res := c.POST("/v1/ca-certificates").WithJSON(cert).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0).Object()
		err.ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.String())
		err.Value("messages").Array().Length().Equal(1)
		err.Value("messages").Array().Element(0).String().
			Equal("additionalProperties 'metadata' not allowed")
	})
	t.Run("creating a certificate with san fails", func(t *testing.T) {
		cert := &v1.CACertificate{
			Cert: goodCACertOne,
			Metadata: &v1.CertificateMetadata{
				SanNames: []string{"XXX"},
			},
		}
		res := c.POST("/v1/ca-certificates").WithJSON(cert).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0).Object()
		err.ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.String())
		err.Value("messages").Array().Length().Equal(1)
		err.Value("messages").Array().Element(0).String().
			Equal("additionalProperties 'metadata' not allowed")
	})
	t.Run("creating a certificate with key_usages fails", func(t *testing.T) {
		cert := &v1.CACertificate{
			Cert: goodCACertOne,
			Metadata: &v1.CertificateMetadata{
				KeyUsages: []v1.KeyUsageType{
					v1.KeyUsageType_KEY_USAGE_TYPE_ANY,
				},
			},
		}
		res := c.POST("/v1/ca-certificates").WithJSON(cert).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0).Object()
		err.ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.String())
		err.Value("messages").Array().Length().Equal(1)
		err.Value("messages").Array().Element(0).String().
			Equal("additionalProperties 'metadata' not allowed")
	})
	t.Run("creating a certificate with expiry fails", func(t *testing.T) {
		cert := &v1.CACertificate{
			Cert: goodCACertOne,
			Metadata: &v1.CertificateMetadata{
				Expiry: 0,
			},
		}
		res := c.POST("/v1/ca-certificates").WithJSON(cert).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0).Object()
		err.ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.String())
		err.Value("messages").Array().Length().Equal(1)
		err.Value("messages").Array().Element(0).String().
			Equal("additionalProperties 'metadata' not allowed")
	})
	t.Run("creates a valid certificate", func(t *testing.T) {
		cert := &v1.CACertificate{
			Cert: goodCACertOne,
		}
		res := c.POST("/v1/ca-certificates").WithJSON(cert).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.Value("cert").String().Equal(cert.Cert)
		body.Path("$.metadata").Object().ContainsKey("subject")
		body.Path("$.metadata").Object().ContainsKey("issuer")
		body.Path("$.metadata").Object().ContainsKey("expiry")
	})
	t.Run("creates a valid certificate with all metadata", func(t *testing.T) {
		cert := &v1.CACertificate{
			Cert: goodCACertFive,
		}
		res := c.POST("/v1/ca-certificates").WithJSON(cert).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.Value("cert").String().Equal(cert.Cert)
		body.Path("$.metadata").Object().Value("subject").Equal("CN=example.com")
		body.Path("$.metadata").Object().Value("issuer").Equal("CN=example.com")
		body.Path("$.metadata").Object().ContainsKey("expiry")
		body.Path("$.metadata").Object().Value("key_usages").Array().Equal(
			[]string{
				v1.KeyUsageType_KEY_USAGE_TYPE_DIGITAL_SIGNATURE.String(),
				v1.KeyUsageType_KEY_USAGE_TYPE_KEY_ENCIPHERMENT.String(),
				v1.KeyUsageType_KEY_USAGE_TYPE_ENCIPHER_ONLY.String(),
				v1.KeyUsageType_KEY_USAGE_TYPE_SERVER_AUTH.String(),
				v1.KeyUsageType_KEY_USAGE_TYPE_CLIENT_AUTH.String(),
			},
		)
		body.Path("$.metadata").Object().Value("san_names").Array().Equal(
			[]string{
				"example.com",
				"www.example.com",
			},
		)
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
				"value '" + digestOne + "': ",
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
			`certificate does not appear to be a CA because ` +
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
	c := httpexpect.Default(t, s.URL)

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
		res.Status(http.StatusCreated)
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
}

func TestCACertificateRead(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.Default(t, s.URL)
	res := c.POST("/v1/ca-certificates").WithJSON(&v1.CACertificate{
		Cert: goodCACertFive,
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
		body.Value("cert").Equal(goodCACertFive)
		body.Value("cert_digest").Equal(digestFive)
		body.Path("$.metadata").Object().Value("subject").Equal("CN=example.com")
		body.Path("$.metadata").Object().Value("issuer").Equal("CN=example.com")
		body.Path("$.metadata").Object().ContainsKey("expiry")
		body.Path("$.metadata").Object().Value("key_usages").Array().Equal(
			[]string{
				v1.KeyUsageType_KEY_USAGE_TYPE_DIGITAL_SIGNATURE.String(),
				v1.KeyUsageType_KEY_USAGE_TYPE_KEY_ENCIPHERMENT.String(),
				v1.KeyUsageType_KEY_USAGE_TYPE_ENCIPHER_ONLY.String(),
				v1.KeyUsageType_KEY_USAGE_TYPE_SERVER_AUTH.String(),
				v1.KeyUsageType_KEY_USAGE_TYPE_CLIENT_AUTH.String(),
			},
		)
		body.Path("$.metadata").Object().Value("san_names").Array().Equal(
			[]string{
				"example.com",
				"www.example.com",
			},
		)
	})
}

func TestCACertificateDelete(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.Default(t, s.URL)
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
	c := httpexpect.Default(t, s.URL)

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
