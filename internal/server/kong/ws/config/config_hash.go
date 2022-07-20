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
	out := make(map[string]string, len(c))
	for k, v := range c {
		out[k] = hashPart(v)
	}

	return hashPart(out), out
}
