package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashPart(t *testing.T) {
	require.Equal(t, "23235a224da3cb921fc8722198f0e76a", hashPart([]string{"test"}))
}

func TestConfigHash(t *testing.T) {
	tinyconfig := DataPlaneConfig{"plugins": []string{"test"}}

	configHash, hashes := getGranularHashes(tinyconfig)
	require.Equal(t, "c96ffe1c067f8f15eb733e540e7ef443", configHash)
	require.Equal(t, map[string]string{
		"plugins":   "23235a224da3cb921fc8722198f0e76a",
		"routes":    emptyHash,
		"services":  emptyHash,
		"targets":   emptyHash,
		"upstreams": emptyHash,
	}, hashes)
}
