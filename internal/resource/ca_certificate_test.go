package resource

import (
	"context"
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

	invalidCertOne = `
-----BEGIN CERTIFICATE-----
MIIE4DCCAsigAwIBAgIBATANBgkqhkiG9w0BAQsFADAaMRgwFgYDVQQKDA9rb25n
X2NsdXN0ZXJpbmcwHhcNMjIwMzA5MDk0MDQ0WhcNMzIwMjI1MDk0MDQ0WjAaMRgw
FgYDVQQKDA9rb25nX2NsdXN0ZXJpbmcwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAw
ggIKAoICAQC5PsnkWbb0/Tg9GuYyTBkWzOUYEfL7kOjUNNp+Kxe85UY5zm5OUjwk
OymG4lr7Foe+VlBQ3DKxAfYFTUOBjMvKL/SnyU7PwZeMY7D97zv7P8BBAvg5NvXr
+ZO349Su4+KdgF9WPGYmJwn/ccSvG1Qugw8ic+DlZXJg06cqr7onkcFchmfUWowr
wC4uUa1igb1y3laFtkHRGbqUliUPg6YOj47EFHBqJyUHv3hyJl5Pq5IEWlS6at79
sRSdW1ITvkD9EP6q/8lWkiqdBnmI04YV1gE38I45WPrzRs7ksYjXzB6HU1MT8gy3
0nUwaQht1cbAMd5yQgRHpMIe8miDRlZiFgECKfz46ynIU8TFBMuzjk6oX0qQfUap
tqQRYfx5G4ejX+jzKncVn98YR4QjHiuUY2W6Gi43k0BO84mWK2sp3GWhCFTrFwrk
ekcbKiIZpJCt2Skfoo6Z5lPGrUJ6Iy+hr9dpAkxpkE022zy7c9sUT8HnQp13Ry49
Xg6C/8B+zWsxQB5ilShkhJ+eqR1ta+SM/CxhQBB991TcRX6Ty/McCbJA6tCtTewx
NKKH8u3YMH4n7ui+g1Ux7AeYL9gXbNESfcB2daP1du2RN/vrmrWBQIywYqYxOBoy
P3YvlRv8owwx4nx9Ir0q/tkd7avLwRRkSs2d7RoxwsglJCzUFCZZtwIDAQABozEw
LzAOBgNVHQ8BAf8EBAMCAoQwHQYDVR0lBBYwFAYIKwYBBQUHAwIGCCsGAQUFBwMB
MA0GCSqGSIb3DQEBCwUAA4ICAQAZWmLockEJGaBzcvxQyEBPrAptxvBs/SXNLJT1
jgsgIHp1Uxx3FGpd6AcDww1S1DBTsEeCyJE2wnmI1Sv+hg/40P4+XvnP+sr5CFrf
KD6UbxhhIBoPQIKAVr849E5w/1E/xyfXUvhK8JxBDcMwjrHenJFVk2yy5xcke762
MTKXRHuaqn0mESOHp8kJ1nGf60tGzSefW1SFvyXm6FRbdgpdqbngmVAZnD5CSv/w
cdvViNDY2tovxW+xMvFJOL7Hu6JspW4PShzMzgz/QOOVmYfKUur/6ym7QQ+dE+9b
uW4IoQ6hhmLisPUf6FyyPL+pRtF05kDUpQApn7062NWSfzZVHB01+Z8W6msf6EfE
S0JRlQ2vPNHzBh6dSnZiCVVt73e5AHJQviNl1WuCgZBSkC+9wAsaOhp66W2IfCKW
Fse4tXi9zithT+1c9TVI6qIN6BXueVapaT2+DwZ4YtuMYd24eunVtfd4bAKPeDn4
AM+NHourg1pQkSdqfaPMlxXHiHihIMDnUU+TA2buOrJ+vB84i5j0u06qSj20+zHe
XAgC4eDa8hIbDB4npueEAwmOTz6/vm+2yBgQsknrcaKyUK6dBwhq7mJqln5sHd1j
kE8xEEqIS+ND090MMgPWS+lnhuLq2ztWwtR8xmZssembJnVCgejsDot5egFDAcYw
R7aUSg==
-----END CERTIFICATE-----
`

	invalidCertTwo = `
-----BEGIN CERTIFICATE-----
MIIE8jCCAtqgAwIBAgICEAQwDQYJKoZIhvcNAQELBQAwgYExCzAJBgNVBAYTAkFV
MREwDwYDVQQIDAhWaWN0b3JpYTESMBAGA1UEBwwJTWVsYm91cm5lMREwDwYDVQQK
DAhmb01NIEx0ZDEbMBkGA1UECwwSZm9NTSBSU0EgQ2xpZW50IENBMRswGQYDVQQD
DBJmb01NIFJTQSBDbGllbnQgQ0EwHhcNMjIwMzMwMDY0NjU3WhcNMjMwNDA5MDY0
NjU3WjAWMRQwEgYDVQQDDAtkZW1vQGxpLmxhbjCCASIwDQYJKoZIhvcNAQEBBQAD
ggEPADCCAQoCggEBAMusWI4QGXU+NH8Iar8qSF5XU08wK7pnKavIa1z2vCSEMtRE
UehLSJuWjn5rDut1HLVcoAcV+kPRdKBNZP56XnQFZHV6p8wBMM6xor8K9B7vknBa
fCz2dX1+rsu3BrBO6QTOaegRdXfvHx9Qk0VHYHvRaM+o47WY7A1dts9yLVpABGCI
1ScC9K0hoCl0Th+jkNHrVPy/1iM623Ws5TCqnr5nnQkKBuDzO/vXc4w7uijCjrXL
TailK+9ol1p1ZCpELSMsrwn9xrzWrNFraNnj9l8z+YletQFMinaBQgMVIfBcEP4x
YS/mVLaJmyZK2IHYrE++mfokyUt8RUXQ9HTbEbMCAwEAAaOB3TCB2jAJBgNVHRME
AjAAMBEGCWCGSAGG+EIBAQQEAwIFoDAzBglghkgBhvhCAQ0EJhYkT3BlblNTTCBH
ZW5lcmF0ZWQgQ2xpZW50IENlcnRpZmljYXRlMB0GA1UdDgQWBBQeax7vnvtoFvr2
9QlRw2RQdVPjKzAfBgNVHSMEGDAWgBR39H6D/oOnso24wy0ldGFWh4CFFDAOBgNV
HQ8BAf8EBAMCBeAwHQYDVR0lBBYwFAYIKwYBBQUHAwIGCCsGAQUFBwMEMBYGA1Ud
EQQPMA2CC2RlbW9AbGkubGFuMA0GCSqGSIb3DQEBCwUAA4ICAQAK4//22q7CEruU
mJPxuQI6yCK2klEA0DkGofChtk2swvEJD5lF23AlT5p4xF2GNxlV8IrtCQKd0bF8
qvH7xc94s7sIAG+XOVSnoItLZFHf6FktBD07m5ktuP9RgcMUiPLDB74Q5Elm3uu3
004LTpYGTQRE2PEuT2Q6SfmuO8+t9fzdgdmCcn1Js4XyvjhUnse/82K4PDw2gSnL
2VSbNgJdXZPHH4bvdLuvtyIXJKLwjBNOSmGTyvLy5NTdxqSx9EMsorAR29ziHEbY
wczSkxtVo9mLy532V52A4KvcG0/nQbg4GbzdGfSmdTwIZ5zOxDJZfyRIPK9npKlc
g2WR3QJ5LYoN8CkeFcc/UPUZgzmQVWVdIWFqKP9WIPCcvDudQaiLPbzPdZXeWxmC
wR9AkhNGUbIsGPwwp7EW8BVVGU6pMAabuaCoxCWctdtlP73LzitYSh8lTah61cyI
ajjd5ATqkYHmqttvEjFL7pwqA0YytBuRNaMPluUCYzHNXswpdW4JtgYF7Cw22JTc
046AnG2thgODeaqYFkoAK4LFQ+PnWsKqM8OGsth3VTlnNDH0NX5pKJ14Yvh0r2rA
jhhv6SAtbw4k9ZJGp/Bsum58u5i/ShPy1NuML3cCjdb/u1/OkWaLhgo1J4L6a7/o
oBjbFPiujznjahpW7JTZsiP0ZiCmJw==
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
	require.Nil(t, cert.ProcessDefaults(context.Background()))
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
				_ = cert.ProcessDefaults(context.Background())
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
			name: "valid certificate, but invalid basic constraint fails",
			CACertificate: func() CACertificate {
				cert := NewCACertificate()
				_ = cert.ProcessDefaults(context.Background())
				cert.CACertificate.Cert = invalidCertOne
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "cert",
					Messages: []string{`certificate does not appear to be a CA because ` +
						`it is missing the "CA" basic constraint`},
				},
			},
		},
		{
			name: "valid certificate, but invalid 'CA' basic constraint is false",
			CACertificate: func() CACertificate {
				cert := NewCACertificate()
				_ = cert.ProcessDefaults(context.Background())
				cert.CACertificate.Cert = invalidCertTwo
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "cert",
					Messages: []string{`certificate does not appear to be a CA because ` +
						`the "CA" basic constraint is set to False`},
				},
			},
		},
		{
			name: "valid but expired certificate fails",
			CACertificate: func() CACertificate {
				cert := NewCACertificate()
				_ = cert.ProcessDefaults(context.Background())
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
				_ = cert.ProcessDefaults(context.Background())
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
				_ = cert.ProcessDefaults(context.Background())
				cert.CACertificate.Cert = goodCertOne
				return cert
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.CACertificate().Validate(context.Background())
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
