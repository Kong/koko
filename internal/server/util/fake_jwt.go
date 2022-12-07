package util

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

const fakeJwtPublicKey = `
-----BEGIN PUBLIC KEY-----
MIIBKDANBgkqhkiG9w0BAQEFAAOCARUAMIIBEAKCAQcC8D7rkI7N434lbvSp8jbw
1GJYoOCgLI4thyxF1VRXoXzLp9F4JGtFU2UADztahKU1CurZxoeaNHGNlEqbgcTd
CjhHYW39zwwKkrALlq+AbmWCd+5eFv/Qe8eNt4PkHn7cIAjCz0OwBcwjtXf2+XWO
Twff6KHm8LPaJ1KqeyY6VegP/Ly8dReS26LpF2hWV9aZ82FnXYS1Y/YnCYygWzQz
KHjBC9ErBffJ77eWHuIBNojjehAvwQDeOmo05io0hjceECQ1XOWtZd6P/nJ+ov4e
PfJllghiOXcxnAJgLCp3S9kjbpgg3O74T+waz6TUKhDVOwNfkQ8qdpCNjkBs8YPC
P7Rnm2JqTQIDAQAB
-----END PUBLIC KEY-----
`

const fakeJwtPrivateKey = `
-----BEGIN RSA PRIVATE KEY-----
MIIEvwIBAAKCAQcC8D7rkI7N434lbvSp8jbw1GJYoOCgLI4thyxF1VRXoXzLp9F4
JGtFU2UADztahKU1CurZxoeaNHGNlEqbgcTdCjhHYW39zwwKkrALlq+AbmWCd+5e
Fv/Qe8eNt4PkHn7cIAjCz0OwBcwjtXf2+XWOTwff6KHm8LPaJ1KqeyY6VegP/Ly8
dReS26LpF2hWV9aZ82FnXYS1Y/YnCYygWzQzKHjBC9ErBffJ77eWHuIBNojjehAv
wQDeOmo05io0hjceECQ1XOWtZd6P/nJ+ov4ePfJllghiOXcxnAJgLCp3S9kjbpgg
3O74T+waz6TUKhDVOwNfkQ8qdpCNjkBs8YPCP7Rnm2JqTQIDAQABAoIBBwFOJDTQ
/o34ClWoZqeCvuLVBGZn979Oa01P6NuQOim+wsdX4RTj4H5n38pZ+bxohVX9Znqb
1CosN3BzOXy/9OlWm88hORFvweKEbAyJv6Vl5FNC4LAMuU8rXGXX6Y8P+LvgwuN9
24w51wbZmdMr1gsDkfTkyd3id5FkvDScBUwOUsmfMdI7KIZ2D737enVza/V8ITa5
g5HSoQrWR5dfZq5GoqlXBNY0p2QpJrrhzOcZuVBB7ca7NY1htUPmPMcH1gMAFJVG
eiykWDDVKaBINWkmO2lCr4qS1DdmP5Ur4OfNnOT6c+yn6pqtdxm48hm9OBGeFM4G
smRuKQl9575JSx0V7md7R2jBAoGEAe7foc1u9yC5jllbjkcI+lxck9AF53OQZz3W
5BB7WknooC7xHu7kUWSGqJbvEa1qUm6RrDCGhKH4OJHd0HyTsuUgVhetaUdnxpYe
zfIj+/ZZqa+xQbptkhmiZW+ZCEOLsi8Ao/9+2ZvSliS9vSUHDbABHOohPiGxL3Pf
Drt/vM7uzjXdAoGEAYUjv3e3WOZIF2oXwEqdBwO7xjCQAhPjEuuvQNh85wdX1Ip8
aG1KoyTme7skgZRSSxo2+F9NdZz1e64aJCQwfj8EfKIEyDeu+5GLvq75zF5rxVr8
Op3loAFaL15479vxjv1NhWuBp//8eEAUl3JqX3DLrgSioBDkEG9Qni25Yo/kTFcx
AoGEAUluD3z9EH/1ZiBwBU+eV2OisTc6pu/UQhX1dk7OfrVSqUd2ddwbm18rERGg
xgjGDWfTi2emNKbJ4YagvYggnmdO1mDerIW/PIB0sy4s7C77Uy1E93dON4LfC111
5v1oAk6tw35yiBPl5NNCh6YdguwWYZQuWvj8xZUB+QGyMBMk/5r1AoGDZ4ANkfLj
I0SzbZVpoK7JSdXsrcfvtYhk5OjVD3+RFyPmNPtH7yG16L+g0zKvgFqu/Qb34qlA
igHE5pavXCzFt08jMxighCb3ZEvN6M4p7Ecv07ZYhNypRRLOnIsACPjjtj2jKefv
XiexeCHB8j2WqvKRk0wJ1NREBsdjevfe3jSzlVECgYNSfm4MmBPVHDMZQe2AY2GT
mR+sMZvP8g+9pN3Tf268Kh4gjHsYWGgwyUFjsI5aIfh7r7lKR+eP/aIWi1qjegTu
B79XfLmuqbGCNiVQjv+1aoAexG2wRj7W0Mzo6VS5DZxYiBo0gbxleJp8oy3qRa84
tueVlk7lQ1AgdVh7MdD//eFbVw==
-----END RSA PRIVATE KEY-----
`

type FakeOrg struct {
	ID   string
	Name string
}

type FakeUser struct {
	ID      string
	Org     *FakeOrg
	IsOwner bool
}

type TokenTimes struct {
	IAT int64
	EXP int64
	NBF int64
}

//nolint:gomnd,forcetypeassert
func GetKAuthBearerTokenHeader(user *FakeUser, tt *TokenTimes) (string, error) {
	if user == nil {
		user = &FakeUser{
			ID: uuid.NewString(),
			Org: &FakeOrg{
				ID:   uuid.NewString(),
				Name: "Fake Org",
			},
			IsOwner: false,
		}
	}
	if tt == nil {
		nowInSeconds := time.Now().Unix()
		tt = &TokenTimes{
			IAT: nowInSeconds,
			EXP: nowInSeconds + (60 * 15), // 15 min
			NBF: nowInSeconds - 60,
		}
	}

	token := jwt.New(jwt.SigningMethodRS384)
	jwtClaims := token.Claims.(jwt.MapClaims)

	jwtClaims["aud"] = []string{"kauth.konghq.com", "foo"}
	jwtClaims["exp"] = tt.EXP
	jwtClaims["iat"] = tt.IAT
	jwtClaims["org_name"] = user.Org.Name
	jwtClaims["jti"] = uuid.NewString()
	jwtClaims["nbf"] = tt.NBF
	jwtClaims["oid"] = user.Org.ID
	jwtClaims["sub"] = user.ID
	jwtClaims["user_is_org_owner"] = user.IsOwner
	jwtClaims["feature_set"] = "stable"

	parsedStr := strings.ReplaceAll(fakeJwtPrivateKey, "\\n", "\n")

	raw, rest := pem.Decode([]byte(parsedStr))
	if len(rest) > 0 {
		return "", errors.New("rest found, Jwt key invalid")
	}
	privKey, err := x509.ParsePKCS1PrivateKey(raw.Bytes)
	if err != nil {
		return "", err
	}
	tokenString, err := token.SignedString(privKey)
	if err != nil {
		return "", err
	}
	return "Bearer " + tokenString, nil
}

var FakeJWTService, _ = New(NewJwtServiceOpts{JwtPublicKey: fakeJwtPublicKey})
