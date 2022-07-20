package config

import (
	"fmt"

	"github.com/cespare/xxhash/v2"
	"github.com/kong/koko/internal/json"
)

const emptyHash = "00000000000000000000000000000000"

func hashPart(v any) string {
	h := xxhash.New()
	e := json.NewEncoder(h)
	if e.Encode(v) != nil {
		return emptyHash
	}
	sum := h.Sum(nil)
	return fmt.Sprintf("%x%x", sum, sum)
}

func getGranularHashes(c DataPlaneConfig) (string, map[string]string) {
	out := map[string]string{
		"routes":    emptyHash,
		"services":  emptyHash,
		"plugins":   emptyHash,
		"upstreams": emptyHash,
		"targets":   emptyHash,
	}
	for k, v := range c {
		out[k] = hashPart(v)
	}

	return hashPart(out), out
}
