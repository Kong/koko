package util

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestJWTEncryptDecrypt(t *testing.T) {
	user := &FakeUser{
		ID: uuid.NewString(),
		Org: &FakeOrg{
			ID:   uuid.NewString(),
			Name: "Test Fake Org",
		},
		IsOwner: false,
	}
	nowInSeconds := time.Now().Unix()
	tt := &TokenTimes{
		IAT: nowInSeconds,
		EXP: nowInSeconds + (60 * 15), // 15 min
		NBF: nowInSeconds - 60,
	}
	header, err := GetAuthBearerTokenHeader(user, tt)
	require.NoError(t, err)
	claims, err := ParseUnverifiedAuthorization(header)
	require.NoError(t, err)
	require.Equal(t, user.ID, claims.Subject)
	require.Equal(t, user.Org.ID, claims.Oid)
	require.Equal(t, user.Org.Name, claims.OrgName)
	require.Equal(t, tt.IAT, claims.IssuedAt.Time.Unix())
	require.Equal(t, tt.EXP, claims.ExpiresAt.Time.Unix())
	require.Equal(t, tt.NBF, claims.NotBefore.Time.Unix())
}

func TestJWTEncryptVerifyDecrypt(t *testing.T) {
	user := &FakeUser{
		ID: uuid.NewString(),
		Org: &FakeOrg{
			ID:   uuid.NewString(),
			Name: "Test Fake Org",
		},
		IsOwner: false,
	}
	nowInSeconds := time.Now().Unix()
	tt := &TokenTimes{
		IAT: nowInSeconds,
		EXP: nowInSeconds + (60 * 15), // 15 min
		NBF: nowInSeconds - 60,
	}
	header, err := GetAuthBearerTokenHeader(user, tt)
	require.NoError(t, err)
	claims, err := FakeJWTService.ParseAuthorization(header)
	require.NoError(t, err)
	require.Equal(t, user.ID, claims.Subject)
	require.Equal(t, user.Org.ID, claims.Oid)
	require.Equal(t, user.Org.Name, claims.OrgName)
	require.Equal(t, tt.IAT, claims.IssuedAt.Time.Unix())
	require.Equal(t, tt.EXP, claims.ExpiresAt.Time.Unix())
	require.Equal(t, tt.NBF, claims.NotBefore.Time.Unix())
}
