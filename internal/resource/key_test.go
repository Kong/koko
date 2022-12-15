package resource

import (
	"context"
	"testing"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/stretchr/testify/require"
)

const (
	pemTestPublicKey = `-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA2gujMwJavJnU9VA3U+RM
fKJAUvcptlncXSA0jJqTU1PNrK6vJzDbmmaNGC7L4hmue2im4fzujc0lYQM4AYBO
N/OVpbQs7zRBijMARiUZUAhUmQDPBNiazjIxh3ETIXOOYuNInGfeWiu1TaPraOss
Vx0gm9BD5O9af/meuBAq1QhwrV3gNxvfNvFAKMxHLiTFImPIXzct/7FrLyxjb1Uw
g14INW+ioNz7Qh8aVdO9XxfLo0mVD3sAsonf7+q0bxfvwvbAy7IWZCVijZdkiFB1
ycYDsNtZ6xWk00dXARM+q3EnWXNKcCIbMSb4OZjIyAudQ9pp/V2hJF9dWZZmZDOo
K6h3K2tYGQfrzD0ANlbRM+G6uS9yPaM5+aL9m8mH2w4ShwsJksp0QF0GMKNYhOl2
0Fcbp7IlegexF/4ZANWehs3/2TQP72P+fDGvheqZf+2fQ3tBGdoBIeHIW2jxIeh4
eaoMLG5WcAmPGVFK0bMC7eljXHSAmVb8kTO9/+hH5jz4GGgr885BB8suOdlM/g69
ZCjH7Wj6eKnaS6oaN/xnxXhL/LwijWA35vGDzF2lBfWTV/VHI98pBtOA+7r5qixd
L5prbjUUvusTCnyT24G2UlVlyNOI1qX5WstAmlqDuQTJI6xQFkhe+fXln205rDO3
B60cbIexAfjSnMk+rEwlWQ0CAwEAAQ==
-----END PUBLIC KEY-----`
	pemTestPrivateKey = `-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDRe6zuZYrcNisP
429laTdUxSKXigd8Hdkje2jDCvlaVg72DeBj5VarYLRkuL/aeBNxfDbtBBAPh2oO
w0dO6cSLET9opHZxJefwzaSVa4vU6pSKQIT7MT5dTvH4FVDwVxUhD/LV6WB1LZMn
NJcFhokK4+lvyVFB+UIED2uRMB/H2Ilf4L+5hwM7PSxZebNye/34Qd4R3BhTrPnB
osk9WbXO+/jYoCaFOzzLCxPnGUlPxQlfyPU1lTQUQP9LL/t7hxLNn+SuKIarb5XA
Ob8ai8Wgw7eORcTx2wlSr5ZsP/Q3ldlxgfSVl0F78Ra8b2Lne7jM9gRW5cG31xHR
F7FR3t+xAgMBAAECggEBAKfp7rAZDLl/Yf0WXVB4ijWU3ymBJobCli7u2QaeYUmb
+doZPWhViKdOmMqznHVOEqfA3XYW75jC/qxes2X50+V1KdKDIb2ImOZYsDhlQGym
q/I1zWJcEpVQlnw4+evsoa8izY/RxdOneHDQos13DZqBHbjRMiUj21rN0XdLj+3r
nNF55or8VAFF4oeNy01dBuSJ2L+eO/kcEyge5ywhmn1Dwp6DYGORGdrnJzaWu2Pa
DzYwc6SJ7svGBsjL5t+4BhoeljpWm0STmZVkL5TUscGoHfNn2jzYyJYQrASbMLGz
QD72XQtrJI7G99B2lAXnf5fecB4wHcXTHvVE7YqUaQECgYEA5/zdoJ64GmkB58S9
fT84g9aTyVDk316LrchnC0wENkG7Fitp6RbhmAA6q01dPppilnknhinTMlBQHpZj
GlrWpkFXhfoaj5jKqbr/ZPT8HUHM6LzHhRWHvN2+9g+4t34iejMdxnLSmMjuvpoD
+MmaBD+wNhVn047Oz1H9SqfxPgkCgYEA5yp9NI/jUsvUIcx2uaUUtmBdkbcqbqnj
JvSkeww/9yBp6yUaVN3clwe+cgHRKZeXpkNwFp8TzWE0mmG1ZGAG4rQQdnDvvktG
zL+/JwMOsx6b5R7aA5DVkAZUykjOKluXvHyjjTGj7jiToa8RjMrOKoadnwWC3Llk
ZhauKbFEfmkCgYEAnzPSKHsj3sP3UcWbQIuVTiyAiSRhnMS2WJFx3bfSICXlrSYn
7ZUNRhHKMWrLNb4fMCJ+tDyZuiqRgRw1cI2sRrYKyV/EwIzbb7VrtS3GopFYfNOo
nLUUzNDkTtqlKg9+u5u+sER2L/GcneL2HNLFRmsqk0MHWJDlbjNW/tfX33kCgYAx
I2QIB0oQMInAQYE/RysW9XcOYXwgl/ZUMo7AJUN3malKNdHaFmsso5XFEEPQ7otq
6UzrUhdYggA3jOuNEaiFCjexpaIgtkmvflb4yPqX8rq6wosfVOtAuUfO1BkXAe9I
PspZWiL5oYcoSFmXrwiSG5ln0zkVCEeiN9H/xNHFeQKBgBfCRd5Hso0iAgiwUS/T
OSSGmeqEIC8Krk/G0V9iw4i9OQoBTFeFrja/3JGqUzvoXHJW012MiJ5ErY1bK8du
Z2uSW/FjsT7HM69XuC0ibPNJ+5Cw7iQJ6QIXTjID3dhv4NywmENhJW71nSyg/RPT
xV73bGApHGnNU8lCGx/9s1dL
-----END PRIVATE KEY-----`
)

func TestNewKey(t *testing.T) {
	k := NewKey()
	require.NotNil(t, k)
	require.NotNil(t, k.Key)
}

func TestKey_ID(t *testing.T) {
	var k Key
	id := k.ID()
	require.Empty(t, id)

	k = NewKey()
	id = k.ID()
	require.Empty(t, id)
}

func TestKey_Type(t *testing.T) {
	require.Equal(t, TypeKey, NewKey().Type())
}

func TestKey_ProcessDefaults(t *testing.T) {
	k := NewKey()
	err := k.ProcessDefaults(context.Background())
	require.NoError(t, err)
	require.True(t, validUUID(k.ID()))

	// k.Key.Id = ""
	// k.Key.CreatedAt = 0
	// k.Key.UpdatedAt = 0
	// require.Equal(t, k.Resource(), &v1.Key{})
}

func goodKey() Key {
	k := NewKey()
	k.ProcessDefaults(context.Background())
	return k
}

func TestKey_Validate(t *testing.T) {
	tests := []struct {
		name    string
		Key     func() Key
		wantErr bool
		Errs    []*v1.ErrorDetail
	}{
		{
			name:    "empty key isn't valid",
			Key:     NewKey,
			wantErr: true,
			Errs: []*v1.ErrorDetail{
				{
					Type: v1.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'id', 'kid'",
						"Keys must be defined either in JWK or PEM format",
					},
				},
			},
		},
		{
			name:    "need some content",
			Key:     goodKey,
			wantErr: true,
			Errs: []*v1.ErrorDetail{
				{
					Type: v1.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"Keys must be defined either in JWK or PEM format",
					},
				},
			},
		},
		{
			name: "key with jwk and without kid isn't valid",
			Key: func() Key {
				k := goodKey()
				k.Key.Kid = ""
				k.Key.Jwk = "xxx"
				// k.Key.Jwk = &v1.JwkKey{}
				return k
			},
			wantErr: true,
			Errs: []*v1.ErrorDetail{
				{
					Type:  v1.ErrorType_ERROR_TYPE_ENTITY,
					Field: "",
					Messages: []string{
						"missing properties: 'kid'",
					},
				},
			},
		},
		{
			name: "key with valid jwk is valid",
			Key: func() Key {
				k := NewKey()
				k.Key.Jwk = "xxx"
				k.ProcessDefaults(context.Background())
				return k
			},
		},
		{
			name: "key with invalid pem is not valid",
			Key: func() Key {
				k := NewKey()
				k.Key.Pem = &v1.PemKey{
					PublicKey:  "1234",
					PrivateKey: "hunter2",
				}
				k.ProcessDefaults(context.Background())
				return k
			},
			wantErr: true,
			Errs: []*v1.ErrorDetail{
				{
					Type:     v1.ErrorType_ERROR_TYPE_FIELD,
					Field:    "pem.private_key",
					Messages: []string{"'hunter2' is not valid 'pem-encoded-private-key'"},
				},
				{
					Type:     v1.ErrorType_ERROR_TYPE_FIELD,
					Field:    "pem.public_key",
					Messages: []string{"'1234' is not valid 'pem-encoded-public-key'"},
				},
			},
		},
		{
			name: "key with valid pem is valid",
			Key: func() Key {
				k := NewKey()
				k.Key.Pem = &v1.PemKey{
					PublicKey:  pemTestPublicKey,
					PrivateKey: pemTestPrivateKey,
				}
				k.ProcessDefaults(context.Background())
				return k
			},
		},
		{
			name: "key with both jwk and pem is not valid",
			Key: func() Key {
				k := NewKey()
				k.Key.Jwk = "xxx"
				k.Key.Pem = &v1.PemKey{
					PublicKey:  pemTestPublicKey,
					PrivateKey: pemTestPrivateKey,
				}
				k.ProcessDefaults(context.Background())
				return k
			},
			wantErr: true,
			Errs: []*v1.ErrorDetail{
				{
					Type:     v1.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{"Keys must be defined either in JWK or PEM format"},
				},
			},
		},
		{
			name: "repeated tags isn't valid",
			Key: func() Key {
				k := NewKey()
				k.Key.Jwk = "xxx"
				k.Key.Tags = []string{"A1", "X3", "A1"}
				k.ProcessDefaults(context.Background())
				return k
			},
			wantErr: true,
			Errs: []*v1.ErrorDetail{
				{
					Type:     v1.ErrorType_ERROR_TYPE_FIELD,
					Field:    "tags",
					Messages: []string{"items at index 0 and 2 are equal"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.Key()
			err := k.Validate(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.Errs != nil {
				verr, _ := err.(validation.Error)
				require.ElementsMatch(t, tt.Errs, verr.Errs)
			}
		})
	}
}
