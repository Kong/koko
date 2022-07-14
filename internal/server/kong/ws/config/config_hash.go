package config

import (
	"crypto/md5" //nolint: gosec
	"fmt"

	"github.com/kong/koko/internal/json"
)

const emptyHash = "00000000000000000000000000000000"

type granularHashes struct {
	config    string
	routes    string
	services  string
	plugins   string
	upstreams string
	targets   string
}

func hashPart(config DataPlaneConfig, name string) string {
	part, ok := config[name]
	if !ok {
		return emptyHash
	}

	h := md5.New() // nolint: gosec
	e := json.NewEncoder(h)

	if e.Encode(part) != nil {
		return emptyHash
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func getGranularHashes(c DataPlaneConfig) granularHashes {
	out := granularHashes{
		config:    emptyHash,
		routes:    hashPart(c, "routes"),
		services:  hashPart(c, "services"),
		plugins:   hashPart(c, "plugins"),
		upstreams: hashPart(c, "upstreams"),
		targets:   hashPart(c, "targets"),
	}

	h := md5.New() // nolint: gosec
	h.Write([]byte(out.routes))
	h.Write([]byte(out.services))
	h.Write([]byte(out.plugins))
	h.Write([]byte(out.upstreams))
	h.Write([]byte(out.targets))
	out.config = fmt.Sprintf("%x", h.Sum(nil))

	return out
}
