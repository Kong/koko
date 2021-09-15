package admin

import (
	"context"
	"testing"

	"github.com/kong/koko/internal/log"
	"github.com/stretchr/testify/assert"
)

func TestMetaService_GetVersion(t *testing.T) {
	s := MetaService{Logger: log.Logger}
	resp, err := s.GetVersion(context.Background(), nil)
	expectedVersion := "dev"
	assert.Equal(t, resp.Version, expectedVersion, "unexpected version")
	assert.Nil(t, err, "expected no error")
}
