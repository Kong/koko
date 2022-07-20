package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashPart(t *testing.T) {
	require.Equal(t, "5e2a17eac1fa241e5e2a17eac1fa241e", hashPart([]string{"test"}))
}

func TestConfigHash(t *testing.T) {
	tinyconfig := DataPlaneConfig{"plugins": []string{"test"}}

	configHash, hashes := getGranularHashes(tinyconfig)
	require.Equal(t, "4fe09342df7064a94fe09342df7064a9", configHash)
	require.Equal(t, map[string]string{
		"plugins":   "5e2a17eac1fa241e5e2a17eac1fa241e",
		"routes":    emptyHash,
		"services":  emptyHash,
		"targets":   emptyHash,
		"upstreams": emptyHash,
	}, hashes)
}
