package util

import (
	"crypto/x509"
	stdjson "encoding/json" //nolint: depguard
	"encoding/pem"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IJwtService interface {
	JwtPublicKey() interface{}
	ParseAuthorization(authorization []string) (*JWTClaims, error)
}

type JWTClaims struct {
	jwt.RegisteredClaims
	Oid        string `json:"oid"`
	OrgName    string `json:"org_name"`
	OrgOwner   bool   `json:"user_is_org_owner"`
	FeatureSet string `json:"feature_set"`
}

type JwtService struct {
	jwtPublicKey interface{}
}

type NewJwtServiceOpts struct {
	JwtPublicKey string
}

func New(opts NewJwtServiceOpts) (*JwtService, error) {
	// parsing in case of bad env variable parsing
	parsedStr := strings.ReplaceAll(opts.JwtPublicKey, "\\n", "\n")

	raw, rest := pem.Decode([]byte(parsedStr))
	if len(rest) > 0 {
		return nil, errors.New("rest found, Jwt key invalid")
	}
	rsaPublicKey, err := x509.ParsePKIXPublicKey(raw.Bytes)
	if err != nil {
		return nil, err
	}
	return &JwtService{
		jwtPublicKey: rsaPublicKey,
	}, nil
}

func (s *JwtService) JwtPublicKey() interface{} {
	return s.jwtPublicKey
}

var errNoClaim = errors.New("unable to parse claims")

func (s *JwtService) ParseAuthorization(tokenString string) (*JWTClaims, error) {
	tokenSlice := strings.Split(tokenString, " ")
	if len(tokenSlice) < 2 { //nolint: gomnd
		return nil, status.New(codes.Unauthenticated, "unsupported Auth method").Err()
	}
	if tokenSlice[0] != "Bearer" {
		return nil, status.New(codes.Unauthenticated, "unsupported Auth method").Err()
	}

	parser := jwt.NewParser(jwt.WithJSONNumber())
	token, err := parser.Parse(tokenSlice[1], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("invalid token")
		}
		return s.jwtPublicKey, nil
	})
	if err != nil {
		return nil, status.New(codes.Unauthenticated, "failed to authenticate").Err()
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		jwtClaims, err := getJWTClaim(claims)
		if err != nil {
			return nil, err
		}
		return jwtClaims, nil
	}
	return nil, errNoClaim
}

// ParseUnverifiedAuthorization Do not use this unless you know what you are doing.
func ParseUnverifiedAuthorization(tokenString string) (*JWTClaims, error) {
	tokenSlice := strings.Split(tokenString, " ")
	if len(tokenSlice) < 2 { //nolint: gomnd
		return nil, status.New(codes.Unauthenticated, "unsupported Auth method").Err()
	}
	if tokenSlice[0] != "Bearer" {
		return nil, status.New(codes.Unauthenticated, "unsupported Auth method").Err()
	}
	parser := jwt.NewParser(jwt.WithJSONNumber())
	token, _, err := parser.ParseUnverified(tokenSlice[1], jwt.MapClaims{})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		jwtClaims, err := getJWTClaim(claims)
		if err != nil {
			return nil, err
		}
		return jwtClaims, nil
	}
	return nil, errNoClaim
}

func getJWTClaim(claims jwt.MapClaims) (*JWTClaims, error) {
	jwtClaims := &JWTClaims{}
	if val, ok := claims["aud"]; ok {
		if aud, ok := val.([]interface{}); ok {
			saud := make([]string, len(aud))
			for i, aval := range aud {
				if sval, ok := aval.(string); ok {
					saud[i] = sval
				}
			}
			jwtClaims.Audience = saud
		}
	}
	if val, ok := claims["sub"]; ok {
		if sub, ok := val.(string); ok {
			_, err := uuid.Parse(sub)
			if err != nil {
				return nil, err
			}
			jwtClaims.Subject = sub
		}
	}
	if val, ok := claims["exp"]; ok {
		if exp, ok := val.(stdjson.Number); ok {
			if intVal, err := exp.Int64(); err == nil {
				jwtClaims.ExpiresAt = &jwt.NumericDate{
					Time: time.Unix(intVal, 0),
				}
			}
		}
	}
	if val, ok := claims["iat"]; ok {
		if iat, ok := val.(stdjson.Number); ok {
			if intVal, err := iat.Int64(); err == nil {
				jwtClaims.IssuedAt = &jwt.NumericDate{
					Time: time.Unix(intVal, 0),
				}
			}
		}
	}
	if val, ok := claims["nbf"]; ok {
		if nbf, ok := val.(stdjson.Number); ok {
			if intVal, err := nbf.Int64(); err == nil {
				jwtClaims.NotBefore = &jwt.NumericDate{
					Time: time.Unix(intVal, 0),
				}
			}
		}
	}

	if val, ok := claims["jti"]; ok {
		if jti, ok := val.(string); ok {
			_, err := uuid.Parse(jti)
			if err != nil {
				return nil, err
			}
			jwtClaims.ID = jti
		}
	}

	if val, ok := claims["iss"]; ok {
		if iss, ok := val.(string); ok {
			jwtClaims.Issuer = iss
		}
	}

	if val, ok := claims["oid"]; ok {
		if oid, ok := val.(string); ok {
			_, err := uuid.Parse(oid)
			if err != nil {
				return nil, err
			}
			jwtClaims.Oid = oid
		}
	}

	if val, ok := claims["org_name"]; ok {
		if orgName, ok := val.(string); ok {
			jwtClaims.OrgName = orgName
		}
	}

	if val, ok := claims["user_is_org_owner"]; ok {
		if orgOwner, ok := val.(bool); ok {
			jwtClaims.OrgOwner = orgOwner
		}
	}

	if val, ok := claims["feature_set"]; ok {
		if featSet, ok := val.(string); ok {
			jwtClaims.FeatureSet = featSet
		}
	}
	return jwtClaims, nil
}
