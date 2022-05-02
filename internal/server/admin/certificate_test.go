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
	goodCertOne = `-----BEGIN CERTIFICATE-----
MIIC5DCCAcygAwIBAgIBATANBgkqhkiG9w0BAQsFADAaMRgwFgYDVQQKDA9rb25n
X2NsdXN0ZXJpbmcwHhcNMjIwMzI1MTAzODMwWhcNMjIwMzI2MTAzODMwWjAaMRgw
FgYDVQQKDA9rb25nX2NsdXN0ZXJpbmcwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
ggEKAoIBAQC27dYTGEGXFRHiZcHzkFNLkU8iCeRIeuSZQDUzt+XykCawoM9oOLx+
Ej/jUwDeknEp9ehaMJpg7AYXFyEGV8Txy5oG5ysdQo9a3559JmtBYLdw3+sAHHRU
w/zITnjqnzURGAiIFF7jLxOfE0ttRXFp/eF7NaOHxMU+ozvbpdHRzDzGrZ3+I/by
d92Vsoz6GWYOHz6Pb1N0AO3PHT5X/OLtuN/ynJowW+r6hjblORVrjAmBjZV5doe4
1kk6rBnwOJa++RQvXRucd37zwfncfkop7NZ4h2KL+5mnUPhsJfQxyhK3V6A2f+LP
Z9U5UgrBaiIwvbIpMmF3NEi/nrO6dj0dAgMBAAGjNTAzMA4GA1UdDwEB/wQEAwIF
oDATBgNVHSUEDDAKBggrBgEFBQcDATAMBgNVHRMBAf8EAjAAMA0GCSqGSIb3DQEB
CwUAA4IBAQAJzRxzUP5rKz4DbeyJePOs6krhnVNnD1ygCa6EX0GuXpyAdGHz4+zI
HRnL3c/01soUZi/jwWrewFkYyucNMQP86IL2McGfTR9XMrIaZAXDCeCDhzegiMTX
2X1qzcUvKgmSCr2WNWMMiWTVsoYmIizmX/d3Ca0+1H1zWpXum0jC5xuGA2ahm8/q
VnK0mafgnc7C0Ts7DUh+rXmS6E+ps7q+VFn0cVhU/NLPpZfACYNqJf3FS0HMEUr7
HnmqVD6SGwH28hX3CDm57EHs7uJIB+ViPzIwyymllbWTcOxRm0IWnpP0/7IFIpWP
VbK8Ti+tF9xX9DGhXEz7Q2OiWGfWIgZT
-----END CERTIFICATE-----`

	goodKeyOne = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC27dYTGEGXFRHi
ZcHzkFNLkU8iCeRIeuSZQDUzt+XykCawoM9oOLx+Ej/jUwDeknEp9ehaMJpg7AYX
FyEGV8Txy5oG5ysdQo9a3559JmtBYLdw3+sAHHRUw/zITnjqnzURGAiIFF7jLxOf
E0ttRXFp/eF7NaOHxMU+ozvbpdHRzDzGrZ3+I/byd92Vsoz6GWYOHz6Pb1N0AO3P
HT5X/OLtuN/ynJowW+r6hjblORVrjAmBjZV5doe41kk6rBnwOJa++RQvXRucd37z
wfncfkop7NZ4h2KL+5mnUPhsJfQxyhK3V6A2f+LPZ9U5UgrBaiIwvbIpMmF3NEi/
nrO6dj0dAgMBAAECggEAVvdpX/iXLjGZKA4CkD3cK7/wZBlZy0+JoIYTBPx3uMLp
ce1xzXWzvygD8ZoDfs0WOcGr7jzPGCb9mjqnu7E8c0u9dWyvZaDAMI7BdXQvZ4yI
iYQa4BmnAKmQYtZTzA9WlkLbw34TwmQeKvFsWY27Jo3Jhd7xWNmmgGnwSjNiNh/Z
BLdUlQjjhfor0WIo/UYnoUN9KTHfN+shd6oTtbOxwhIsd+x2FnN23NbVzdbICxXN
2rKQAeOt2Kt0ouCcsSqXPv4DDhuWEdgAwXuEwrUDgw2NuxT3bPkgJV50LWsFvvSU
FkUfjFprktPzs6bZj2vB4WFvqyDXgHUa8RxdecgbjQKBgQDCD7X/0bdsJtq2/gk4
Zth/bhInaRVN7DJVUmLSUhixC+/arXDEk41IaDI8hJBHfmm0cyTEJChlRiYjsOV0
vPYCRx1ibfDIxaO2/+s7OBhMQGHR57p+zL8OV9yPW/M73wVxYP/rrlbDKMyXyrjc
7WGD1Rd6KJ/yPPfErVvox0Pu2wKBgQDxUIemWqdzFXMNJDqLfYtiQTUomI9gN5p1
hyRnJO4IzThTDfconag6kLQjWw8mRUxC5NB4fkRNOFDf1dVn1Vf9F3ohvqxvY0dZ
R/cAIB2JkRenBB7pCfXqIeO7CjTplt8gwERyhFk0+HAdnOsN4WYzlYX8NvXybOLF
xK/WMnVZZwKBgQCFJ5KRvaxFoUNhXF2nPao/hZ8fO5NKrE69DJKSDZKzqKUjPu6p
czT2Aci2jZ8R70NIddk8XDL7im6Q/sfymdWTKoiXCSi2GiaDYoZdU9gYOfTkukPU
zVgq106Xb1guNJDfgtcXN8CAmHYJkSfXL2pBsu0w/L8Cz6KSaQEvb8rFEQKBgGiw
lykEi0DSPWemIAAEJ0QpJfbGuOz1Mn1qc9CLpPkMjzL5DBEC1MkTnhL4nknsJnme
6xJbNSaLGAsDqeGyHMogNUwOfKCWYY3KOs5DII1d9PTwRLi1KYq5ySKL+wib+5Ep
2IgWAt2IKpuuSAttjfkzFT2mWm0h7//8pIw4t9BhAoGAGZMTDWtRwMWpb0PTtUNs
KeclaWPHzsLG1zdMXEoop7eBTDioXtkvZvQLXjtNUenm2vYLei3+CxyMj+l5woF7
sbBr379NQCHO3UOO3cbje82D2ocE3JFPCtp2odJZGvInIGbCu9ClnHJK48TnHINh
L/7B6a3oihBkiJmT4tR2DNo=
-----END PRIVATE KEY-----`

	goodCertTwo = `-----BEGIN CERTIFICATE-----
MIIC5DCCAcygAwIBAgIBATANBgkqhkiG9w0BAQsFADAaMRgwFgYDVQQKDA9rb25n
X2NsdXN0ZXJpbmcwHhcNMjIwMzI1MTAzODMwWhcNMjIwMzI2MTAzODMwWjAaMRgw
FgYDVQQKDA9rb25nX2NsdXN0ZXJpbmcwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
ggEKAoIBAQDrYbW8y+3GkGykVWzrwnyhqiiqrplNCdP3S9O6d56c4eMVj2/hryp9
9i7d1WdNFDv8/LXVBhwPB6pyUCYaGd6ZUPYN4VEdsXvD9KbXrMZ738KfZC6ikSgx
EqqBV3sQqSoAAmNXHsN30z14+1rHU0cGSznW8MCd6KTfwTJSg4vGinv2eEegz3Fh
qG8Vp3MmFY9KqUo3ePDxcCr6MiTs+u0gC2YOM9DV+AM3fr9EOAWyuDLIbf2zkxFl
Ek+Ll80tq4RvQptAcdYdbnqerzAXdnwbmxEBYeG4kdm/Yc4i6JHRGQggWwN2KJ7h
TwLvMC5gvZIGqkBrTXJyfDGlUt+hJ/lXAgMBAAGjNTAzMA4GA1UdDwEB/wQEAwIF
oDATBgNVHSUEDDAKBggrBgEFBQcDATAMBgNVHRMBAf8EAjAAMA0GCSqGSIb3DQEB
CwUAA4IBAQBE3N+pl0rwS8sszWbxXDP3f+TsVwoZ6f1Vm5uZHRJPS5mFOta28ojw
m5t/d8DV0Fy2NEQGt4SIV2p2K/p6fKHnE5ncJbxrtqWmpHawRYbOADQloBkLqNd5
2Vx56ytzROrm+Q22GuUlrB2RW/aJMWhiu8GD9vUwUrGXsxaWMA8KvPRRT+93Z8s1
yUZ/nzKAXtvTG1GBRKHmVpPF9YS8RAlPlr3r6fFDC864YTXVEJSfozKQlYn5ve7M
7MyACeB3IaGN69XXbn45lthx8CYdFCUvSlNxfcRXyYoWyyrIRArZ9PXApuIzMu4D
mS1kv4+P6RyW3lL7bU+/pNIATQ+9AONw
-----END CERTIFICATE-----`

	goodKeyTwo = `-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDrYbW8y+3GkGyk
VWzrwnyhqiiqrplNCdP3S9O6d56c4eMVj2/hryp99i7d1WdNFDv8/LXVBhwPB6py
UCYaGd6ZUPYN4VEdsXvD9KbXrMZ738KfZC6ikSgxEqqBV3sQqSoAAmNXHsN30z14
+1rHU0cGSznW8MCd6KTfwTJSg4vGinv2eEegz3FhqG8Vp3MmFY9KqUo3ePDxcCr6
MiTs+u0gC2YOM9DV+AM3fr9EOAWyuDLIbf2zkxFlEk+Ll80tq4RvQptAcdYdbnqe
rzAXdnwbmxEBYeG4kdm/Yc4i6JHRGQggWwN2KJ7hTwLvMC5gvZIGqkBrTXJyfDGl
Ut+hJ/lXAgMBAAECggEAFbuQzxy5GINPNKEajG4JmdefJ5s1VlKY+pVKuEXBJFUK
5Xu35cuJjdXEIAFLJ2e93i7rDv8gahbYGvPhgLlwvxEllsR9+9LeTYpyOSmfreJM
EvFqCOKEJnvFuGl+WFx2H0gZKfsSKyca3ue+SvcacK9TaNATyMCpMGujHNE+f5cu
YLrGU6PXj04GbXAG7dV+7ciaooDx5+g5WXkczd4DBJdBYRsiN0qTgpSPLJozJdBp
3YExba+QyH7Kf64+hOlwlmLGPeelmBYipRRQatP/6EDcVoxNAwuHmTu5RqPBHLpA
YqH2hQbj0DRP69F83MaqZMKimYUV0d379mxQ+nwaYQKBgQDvUe3EpsqzXaUL+Q6m
ezbnrUWpT8wkZDqImDzlk6b4ZUQlpD73InAU2gnrxCGG+7P4NijdxyWxg77QBXQe
VLRgnFp1RgMSNsOTd7CJnKnEOslwq3tTknHSPMo5k4CbbRUg+lit0dRZrU7SrVs6
QqRHfaAWS0ZRKz6RSMCXk9ddhwKBgQD7yYLPMkefCmtN+xzd2ZnfnhX6Neskbfs9
XcADcWiQeotuQmuCEL4CB/4/+AM2qlfMfer4oBsGcMPdrfYRShigQZvtQhAwF2qx
M6VyLqEcY5mLYFLLm0Ac+EbWP2b1JLVgUh7iptmr1ggNY/+nqFX+nlGl84i+k/Oj
eZ/jKHj5sQKBgQDbjqGBYbfjOI1734GGUNI9WCTpwTC7Tky5FloAEScFCfqsQfQW
TLzhFGw3pZdQvEkO9bkmRlcZdZGwTOCMFw/o9miy4Ilew2lIOG14woapZXl7aYda
U6cixuyMR/ucHEZfG+4Rgci8gRgohiyE1bDbebBN479eJjtflIxEQ7k4rwKBgF43
Eh6L1ub5FBvy6eNNyFk3o0ukH1/bU5ar3Oys5A6j/EZ+zhG2SBMkgIvZNwKejQn0
2Ba+ej5XtcLelGP10O8ufbUy8jG8oWy7QZ5POnQQBOV1XqXXaw8sC/2hbdovKTto
nyv6eRrmlM7F62UGBV+oSC8LyNBfNlymZyCuBU6RAoGBAOAJnsKD3gwPthDy6uor
czqLkIw7zKO+dGyGzxPOU/exumB/BZoGI2l+Z0fYrA8BwprJl1wk1zJiHnD2GOMC
pSpGvBjcg9PalMTNajHPNYnFNeEI+uoB4zRgmxSJKjKhpZs7QliqaMDE2Uqttb80
7CGu2Xy4pJRnu36kWhdBzExx
-----END PRIVATE KEY-----`

	goodCertThree = `-----BEGIN CERTIFICATE-----
MIIC5DCCAcygAwIBAgIBATANBgkqhkiG9w0BAQsFADAaMRgwFgYDVQQKDA9rb25n
X2NsdXN0ZXJpbmcwHhcNMjIwMzI1MTAzODMwWhcNMjIwMzI2MTAzODMwWjAaMRgw
FgYDVQQKDA9rb25nX2NsdXN0ZXJpbmcwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
ggEKAoIBAQCxWp+qO+A5umX8o22ybJWTHufn4S0bvYeoUwZPE6BfZ0k1qhDv+XXp
o+sCGU5KSiWIcBoqLdil6RRS64JOz4i8DyXd1CFW+zJPkCSxZXZ2jSJTN5wNakVR
ZmrATnGZH6aN922OBP2gAz3byTyTKnQNuudIrWdeg0nco5zK/tv4AeNPkUJycsXJ
wbifCx5EbIXKsRzSFZhteMqERGHcoXByxsYlfkZR6Z2baNR1ApM1aP6UPIaEuBi+
36Z4kGOBPJ7f+oxZVfMQx0tf6z3cZ1BNcwiC5Qc3Jk/6kxJzdN2SzMc/W0cRTDeG
Y+uDA0ygYIrApHVZsuVWITpQoXQZo2z5AgMBAAGjNTAzMA4GA1UdDwEB/wQEAwIF
oDATBgNVHSUEDDAKBggrBgEFBQcDATAMBgNVHRMBAf8EAjAAMA0GCSqGSIb3DQEB
CwUAA4IBAQB4g4s7+nspmPDcoOFYtWFOupt8J64aEQNO0wQQZtTmU+a6FPXvA8QE
J+Ljv7JEth6roeOtwFbOYDrW7OKbiYEKPqln/PUJy80Xma4NORMgJuz5YQ7WdB7u
QCMwf1pr8Iq5PqPel6s/CZ1CbxSB3yr17SnMxpno1JC1IYNG/sIjRO181sdCCwlN
vF3F05WkuDxi8fRDLmwGe4hQvdL2o4o0oTr/57MXO/1od1YNlb3QvJnGTKKvHsJP
Z2uQ8kf8/kt7m2Rv8UoUrveoXKvGudyWmY0dAOQjWJpUB5E7jgfedIYUMiWhntpU
3ryC4iTGbJpvXh/2zdp0ZMMyxXN3z0sZ
-----END CERTIFICATE-----`

	goodKeyThree = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCxWp+qO+A5umX8
o22ybJWTHufn4S0bvYeoUwZPE6BfZ0k1qhDv+XXpo+sCGU5KSiWIcBoqLdil6RRS
64JOz4i8DyXd1CFW+zJPkCSxZXZ2jSJTN5wNakVRZmrATnGZH6aN922OBP2gAz3b
yTyTKnQNuudIrWdeg0nco5zK/tv4AeNPkUJycsXJwbifCx5EbIXKsRzSFZhteMqE
RGHcoXByxsYlfkZR6Z2baNR1ApM1aP6UPIaEuBi+36Z4kGOBPJ7f+oxZVfMQx0tf
6z3cZ1BNcwiC5Qc3Jk/6kxJzdN2SzMc/W0cRTDeGY+uDA0ygYIrApHVZsuVWITpQ
oXQZo2z5AgMBAAECggEAacEXKiRwDRxICkDNfbJf8o1gTZWpFzyJ8uYnAeo7HAhz
0Csr1FzVYc9bqDG8zHGwNc5a28HgyPXWJ6fFWQdJipIhy0fd5Yb+NhFGv/03iXOY
/zROunUfBm3iw+9Cr8L+xvK+ggwZzFuCfFdf5oVPFIzZsy2rUOFKnuV07lrQge5o
Hfc835i9z1c16EG+xH4XyyLZq9F0G1O2OCrn9FzOrL3R2iAE6d/8Ie0JjqTlKLUh
LXWH4okFZBd39wp4VfsGFmWQVBePhrGC1GUMUyukXqJxk+J6A/NaEALzNfy3tQg3
Kql0vPGaQAQImxWHVyjE+FehIgfGX7Pls5U/mFIyLQKBgQDYe0YNtB3QVBWRfIOE
2khFkaGyRkOXxE7CMeakgaXERBSIbV0dtIUPKXhZwO8TN7/Yii43oApKd27GbTIp
zM7dRXyIGPf1P2xrSnx863qWK0KjqsHgJANsfIZMSNeK/+xcgxkuplLvLokvnhDF
e53U1XE5kfd4++uh9bZBbpRwfwKBgQDRutJiLl+P3LyRUHQLfUBVGVE1HPq6xye0
ENMDCplxTnU/uxCtL0l9A46Fg66lBnuZ9X6nF9/SZhfbBD6VGVySrwSvEr3Hvtya
xspqLjHjIialF7uBifj4Qz/h2ccZZlMt/zJzChq3fmDF5YyWbcF8GbGrA/y+2baG
ptLVjrbmhwKBgQCza3BDOU4gdSAvJYCnonaV2j6mz9+DsLsJ7mvXWnC2Oyq9a87q
KMzDJT7PPL2pMuJ5KQVnKuh8kYIpSSVzSYEGGWo+LluMUxWb0u9cZZqDTbV8irEH
ATIpPwfbv9+NH9GZVzqO1GEWRX6EDcCevHayiHjAGz99cWX5JPn6PxkeJwKBgA22
dlz/BTaFyzqBFSVPKi6mOh4L6ATgUqM+Wl2fisrSw23IUF3Scq6e1em642dc1iYJ
3B8Wu4apMDQcwe7Dur7IfLjps8jknM3t43wvywk7yWUP+S2OFN4+n2Wn9JGflB60
ydqltXt74t3tlVScloMDtw8kcpkT6RBCxhQ6gZDvAoGAECj0L76cKd9wo8StPAep
Pd6MvwCuepFMmKHaCy11x3G/YMX1PFcj/ZQLbjztXjG+EPMtsHr1YPaDqgs21pCj
SzOKUD4gI1Wg8vIvjpfQomLkfNy5e1cWRmhZh23yrNmZ3LQscDzoUlMg+/DP5pyO
QRR8Bmf7kE6sb7KDrouvZzI=
-----END PRIVATE KEY-----`

	goodCertFour = `-----BEGIN CERTIFICATE-----
MIIC5DCCAcygAwIBAgIBATANBgkqhkiG9w0BAQsFADAaMRgwFgYDVQQKDA9rb25n
X2NsdXN0ZXJpbmcwHhcNMjIwMzI1MTAzODMwWhcNMjIwMzI2MTAzODMwWjAaMRgw
FgYDVQQKDA9rb25nX2NsdXN0ZXJpbmcwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
ggEKAoIBAQC+9crW6MtNzpLrytA0Hk1G4q54tLSQjzItuI08K1MQlzKZZnnnFsJ2
TYCkQJWpp2n7hTq4cAWzH+HBunyzI0e8/suPiNqn/exH81gINrwlK05qlxto5Fsy
Et7mW6AUI32o8pNixO9JSX3y4x0R9PZUw2NM5tTE5GQouyOwgkwOHaCt+zaX5dhD
zXcwxF6wntatPWHKz/nfQyKBXzmepbx1Z/0YOqe1GUlr8MhlekSpD93yVCTtH/xB
cWvmeEL2OFDxHF/cZrSFUKhi1RT8WNDmVzz2pR3jDrWlUwKO/047xDOUozRZA8uq
IuuuY+d3rBigOI2KWLBDtJlPI3/6HpUPAgMBAAGjNTAzMA4GA1UdDwEB/wQEAwIF
oDATBgNVHSUEDDAKBggrBgEFBQcDATAMBgNVHRMBAf8EAjAAMA0GCSqGSIb3DQEB
CwUAA4IBAQCgss1HHHbNv1Ej56dT7nDO0nz0+Tf+s7WB+WFyG9OA8brbnzQZF5dJ
uOliptLnBjnl4+1vhkHOSMXeCEV//GuF5fZmDdMGunnEIfWq9VWrwxYj7RcK1weF
6RpWSserN3U5zClY5iGWEtOsBP3p0X4fhiLKx+8ortlF94mp9jxM3PjEG3lG1GK+
0w0MCjhIDQnkLQnVv99TRUuzyPLPm5QnWDt0gy3llQd9KEjpcPJ+CeId/EXfXY2Y
R0v+DImlWGjKFVZFmkRcWq/lZs8TsWPSbWrld9S//Dq2zN57rMxNTMYxT7Titmm8
ZN6MQL0X0onFChJXMi1UpNvzH2RsSrZa
-----END CERTIFICATE-----`

	goodKeyFour = `
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC+9crW6MtNzpLr
ytA0Hk1G4q54tLSQjzItuI08K1MQlzKZZnnnFsJ2TYCkQJWpp2n7hTq4cAWzH+HB
unyzI0e8/suPiNqn/exH81gINrwlK05qlxto5FsyEt7mW6AUI32o8pNixO9JSX3y
4x0R9PZUw2NM5tTE5GQouyOwgkwOHaCt+zaX5dhDzXcwxF6wntatPWHKz/nfQyKB
Xzmepbx1Z/0YOqe1GUlr8MhlekSpD93yVCTtH/xBcWvmeEL2OFDxHF/cZrSFUKhi
1RT8WNDmVzz2pR3jDrWlUwKO/047xDOUozRZA8uqIuuuY+d3rBigOI2KWLBDtJlP
I3/6HpUPAgMBAAECggEAIgkrMyzg96hHFZHblD6GZYkHsen7ePyc4/tN6RiLwJxC
X4cdWSv8UxuzPxNn2YpGYJc5hSAqU+ft1BrKGR/DrJL5c0bgOisPDy/3U9d1p2ZV
nrf6IbL58i3c3tAb8xr0TcWWsXcKc1SPB1ilmMrBkRAWReGqsMFIfN4GGXLP3X/p
fhKllJs4lNTg/OkyPf04eNDkJuDrySH487OMhOB44pbnijS+D+WWzAtfRAegv9IA
EPwDxyhWB3BLj+/kPg2DgukrA6hbQBPET/gs7v6KpWv1p8sE5J2ClPxYR5aCXXZi
Xy1UXNVbfo0OMAsu5AyB7VzMG7/8TXoAW9vt+8EmEQKBgQDUGcpfn8VJB/c02PEq
C7Bp3vVMj/XWg7/a2p3Vn6Az/IUouem6wOhiQF3cYwZOcgjvCDjSo594B4jw+ocO
+3zJva9bf3PYx0LcD3VkLuEvMswshWRH7AeTTMqFp36plHt6plZt2k1KLmvqYxN5
vzI7tFPDT6DN12774EGGWZua2QKBgQDme9xtWc+Tdl1D51kyQFlXvmgPGTWk/TEj
Gk6iSzk3UPAbl8Bn9Dh4gLaukgbNqSB9Xk2fqyKPDr+g/ApbzUjDkemWvlAidKKZ
Sgx/+F5IZQyKt4A+1zgRcBKNbi436D+WSpgtH6tuWkfw2uY0GTnkEaOUaexUAMof
T1x9Fe8uJwKBgCzqATbaqHZcn3arcfZLX2Ir+pnp6k8wuxHnNYElOlGH6dLD+8C2
VP9pTfb7aTx3XXjwrse8KmrKfa85/huoGbbG2jlv9eIz3+6lv2AlpT3PbfkHjkLE
sp34pvJCk8npTXSdgLPmhHNu/R83N8qSOFr4RryXQiAUvMXNqVJ/6zmRAoGADcPT
9Evq07m79DQ65X9mVpEukchFpebhKmGF1Ld9YUpaLKuxeAPj4358aoyaD2pMYHBj
XmfQFo8g7rJexADMmbF7K9+N1aD1nQYJHRNuPhCa4SX4aMhdttzknsG3zOr38Tff
QsKjcGG/7iiEmxPumypahKCW1qV9bMVGlsnakP0CgYEAnbwbZpMiM9otCl3lsq5n
foGhxWcWTMs0BKA+LEDN0MZXmg+1Mf/3v3PpLu6PvOiMVUX2JNYxYbdZGQMNgBjJ
fUVx6BoQa3iIRtyGKHZUn1FTVc8JL/n1oQu+SK9wkpgnF2bdUzx9QKHKcObtwgPY
MvNxuAXHfgbJohauyWxpe+Y=
-----END PRIVATE KEY-----`

	goodAltCert = `-----BEGIN CERTIFICATE-----
MIIBWDCB/qADAgECAgEBMAoGCCqGSM49BAMCMBoxGDAWBgNVBAoMD2tvbmdfY2x1
c3RlcmluZzAeFw0yMjAzMjUxMDE2MjVaFw0yMjAzMjYxMDE2MjVaMBoxGDAWBgNV
BAoMD2tvbmdfY2x1c3RlcmluZzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABF1B
TKI4zmhL2KQqN4FcoqkEMUx/Yaa+/SlHWkkTWurGNRf1NYnj71dwNGsx0i21ObVF
DjQaoRjz+QzEvn4pmJajNTAzMA4GA1UdDwEB/wQEAwIFoDATBgNVHSUEDDAKBggr
BgEFBQcDATAMBgNVHRMBAf8EAjAAMAoGCCqGSM49BAMCA0kAMEYCIQDt7NDpeqzx
ILQe1oqaMPJXelN+2c0dS1kFN4YFhpdfcAIhAIaGO3JmL4Yj3JbnGG4ch3t7BtwM
PXMNYgeWHWHKWkCm
-----END CERTIFICATE-----`

	goodAltKey = `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgiyrH72+AVZDBktm8
yE95Ch00zcUjbKJXx0Kvq6wca9ehRANCAARdQUyiOM5oS9ikKjeBXKKpBDFMf2Gm
vv0pR1pJE1rqxjUX9TWJ4+9XcDRrMdIttTm1RQ40GqEY8/kMxL5+KZiW
-----END PRIVATE KEY-----`
)

func TestCertificateCreate(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	t.Run("creates a valid certificate", func(t *testing.T) {
		certificate := &v1.Certificate{
			Cert: goodCertOne,
			Key:  goodKeyOne,
		}
		res := c.POST("/v1/certificates").WithJSON(certificate).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.Value("cert").String().Equal(certificate.Cert)
		body.Value("key").String().Equal(certificate.Key)
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
		certificate := &v1.Certificate{
			Cert: goodCertOne,
			Key:  goodKeyOne,
		}
		res := c.PUT("/v1/certificates/{id}", id).WithJSON(certificate).Expect()
		res.Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.Value("id").String().Equal(id)
		body.Value("cert").String().Equal(certificate.Cert)
		body.Value("key").String().Equal(certificate.Key)
		certificate.Id = body.Value("id").String().Raw()
	})
	t.Run("upsert an existing certificate succeeds", func(t *testing.T) {
		certificate := &v1.Certificate{
			Cert: goodCertTwo,
			Key:  goodKeyTwo,
		}
		res := c.POST("/v1/certificates").WithJSON(certificate).Expect()
		res.Status(http.StatusCreated)
		body := res.JSON().Path("$.item").Object()
		id := body.Value("id").String().Raw()

		certUpdate := &v1.Certificate{
			Id:      id,
			Cert:    body.Value("cert").String().Raw(),
			Key:     body.Value("key").String().Raw(),
			CertAlt: goodAltCert,
			KeyAlt:  goodAltKey,
		}
		res = c.PUT("/v1/certificates/{id}", id).WithJSON(certUpdate).Expect()
		res.Status(http.StatusOK)
		body = res.JSON().Path("$.item").Object()
		body.Value("id").String().Equal(id)
		body.Value("cert").String().Equal(certificate.Cert)
		body.Value("key").String().Equal(certificate.Key)
		body.Value("cert_alt").String().Equal(goodAltCert)
		body.Value("key_alt").String().Equal(goodAltKey)
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

func TestCertificateRead(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	certificate := &v1.Certificate{
		Cert: goodCertTwo,
		Key:  goodKeyTwo,
	}
	res := c.POST("/v1/certificates").WithJSON(certificate).Expect().Status(http.StatusCreated)
	id := res.JSON().Path("$.item.id").String().Raw()
	t.Run("read with an empty id returns 400", func(t *testing.T) {
		res := c.GET("/v1/certificates/").Expect().Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "required ID is missing")
	})
	t.Run("reading a non-existent certificate returns 404", func(t *testing.T) {
		c.GET("/v1/certificates/{id}", uuid.NewString()).
			Expect().Status(http.StatusNotFound)
	})
	t.Run("reading with an existing certificate id returns 200", func(t *testing.T) {
		res := c.GET("/v1/certificates/{id}", id).
			Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.Value("id").Equal(id)
		body.Value("cert").Equal(certificate.Cert)
		body.Value("key").Equal(certificate.Key)
	})
}

func TestCertificateDelete(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	certificate := &v1.Certificate{
		Cert: goodCertTwo,
		Key:  goodKeyTwo,
	}
	res := c.POST("/v1/certificates").WithJSON(certificate).Expect().Status(http.StatusCreated)
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
	certs := []*v1.Certificate{
		{
			Cert: goodCertOne,
			Key:  goodKeyOne,
		},
		{
			Cert: goodCertTwo,
			Key:  goodKeyTwo,
		},
		{
			Cert: goodCertThree,
			Key:  goodKeyThree,
		},
		{
			Cert: goodCertFour,
			Key:  goodKeyFour,
		},
	}
	for _, certificate := range certs {
		res := c.POST("/v1/certificates").WithJSON(certificate).Expect().Status(http.StatusCreated)
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
