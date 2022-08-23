package ws

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateRequest(t *testing.T) {
	t.Run("no parameters, fail", func(t *testing.T) {
		r, err := http.NewRequest("UPGRADE", "http://example.com/v1/outlet", nil)
		assert.NoError(t, err)

		err = validateRequest(r)
		assert.ErrorContains(t, err, "invalid request, missing")
		assert.ErrorContains(t, err, "query parameter")
	})

	t.Run("accept sensible parameters", func(t *testing.T) {
		r, err := http.NewRequest("UPGRADE",
			"http://example.com/v1/outlet?node_id=000000&node_hostname=example.com&node_version=2.8.3", nil)
		assert.NoError(t, err)

		err = validateRequest(r)
		assert.NoError(t, err)
	})

	t.Run("accept longer-than-semantic version", func(t *testing.T) {
		r, err := http.NewRequest("UPGRADE",
			"http://example.com/v1/outlet?node_id=000000&node_hostname=example.com&node_version=2.8.3.2-metapatch-edition", nil)
		assert.NoError(t, err)

		err = validateRequest(r)
		assert.NoError(t, err)
	})

	t.Run("reject too old version", func(t *testing.T) {
		r, err := http.NewRequest("UPGRADE",
			"http://example.com/v1/outlet?node_id=000000&node_hostname=example.com&node_version=2.4.82", nil)
		assert.NoError(t, err)

		err = validateRequest(r)
		assert.ErrorContains(t, err, "unsupported dataplane version")
	})

	t.Run("many builds does not a new version make", func(t *testing.T) {
		r, err := http.NewRequest("UPGRADE",
			"http://example.com/v1/outlet?node_id=000000&node_hostname=example.com&node_version=2.4.82.996-extra-strength", nil)
		assert.NoError(t, err)

		err = validateRequest(r)
		assert.ErrorContains(t, err, "unsupported dataplane version")
	})
}
