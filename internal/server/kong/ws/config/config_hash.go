package config

import (
	"crypto/md5" //nolint: gosec
	"fmt"

	"github.com/kong/koko/internal/json"
)

func configHash(d interface{}) string {
	h := md5.New() // nolint: gosec
	e := json.NewEncoder(h)

	if e.Encode(d) != nil {
		return ""
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}
