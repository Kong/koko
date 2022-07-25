package config

import (
	"fmt"

	"github.com/cespare/xxhash/v2"
	"github.com/kong/koko/internal/json"
)

const emptyHash = "00000000000000000000000000000000"

// hashPart returns a hash for any Go object.
// Result must be consistent for same values, no undefined ordering.
// Must be 32 chars long.
// (xxHash is only 16 hex chars, we duplicate it to match 32 chars.)
func hashPart(v any) string {
	h := xxhash.New()
	e := json.NewEncoder(h)
	if e.Encode(v) != nil {
		return emptyHash
	}
	sum := h.Sum(nil)
	return fmt.Sprintf("%x%x", sum, sum)
}

// getGranularHashes creates a hashing signature for a configuration object.
// Returns a "total" value that represents the complete configuration,
// as well as a map of separate hashes for each part of the configuration.
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
