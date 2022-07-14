package config

import (
	"crypto/md5" //nolint: gosec
	"fmt"

	"github.com/kong/koko/internal/json"
)

const emptyHash = "00000000000000000000000000000000"

func hashPart(v any) string {
	h := md5.New() // nolint: gosec
	e := json.NewEncoder(h)
	if e.Encode(v) != nil {
		return emptyHash
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func getGranularHashes(c DataPlaneConfig) (string, map[string]string) {
	out := map[string]string{
		"routes":    emptyHash,
		"services":  emptyHash,
		"plugins":   emptyHash,
		"upstreams": emptyHash,
		"targets":   emptyHash,
	}
	h := md5.New() // nolint: gosec

	for k, v := range c {
		hp := hashPart(v)
		h.Write([]byte(k))
		h.Write([]byte(hp))
		if _, known := out[k]; known {
			out[k] = hp
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), out
}
