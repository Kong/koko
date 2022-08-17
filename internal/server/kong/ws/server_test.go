package ws

import (
	"net/http"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
)

func TestTrivialVersionParse(t *testing.T) {
	for _, test := range []struct {
		given  string
		wanted semver.Version
	}{
		{
			given:  "2",
			wanted: semver.Version{Major: 2},
		},
		{
			given:  "2.8",
			wanted: semver.Version{Major: 2, Minor: 8},
		},
		{
			given:  "2.10.3",
			wanted: semver.Version{Major: 2, Minor: 10, Patch: 3},
		},
		{
			given:  "2.32.19.12",
			wanted: semver.Version{Major: 2, Minor: 32, Patch: 19},
		},
		{
			given:  "2.32.19.12-superjuiced",
			wanted: semver.Version{Major: 2, Minor: 32, Patch: 19},
		},
		{
			given:  "2.32.19-alfa",
			wanted: semver.Version{Major: 2, Minor: 32},
		},
		{
			given:  "2.12a.4",
			wanted: semver.Version{Major: 2},
		},
	} {
		t.Run("parseversion "+test.given, func(t *testing.T) {
			assert.Equal(t, test.wanted, trivialVersionParse(test.given))
		})
	}
}

func TestValidateRequest(t *testing.T) {
	t.Run("no parameters, fail", func(t *testing.T) {
		r, err := http.NewRequest("UPGRADE", "http://v1/outlet", nil)
		assert.NoError(t, err)

		err = validateRequest(r)
		assert.ErrorContains(t, err, "invalid request, missing")
		assert.ErrorContains(t, err, "query parameter")
	})

	t.Run("accept sensible parameters", func(t *testing.T) {
		r, err := http.NewRequest("UPGRADE",
			"http://v1/outlet?node_id=000000&node_hostname=example.com&node_version=2.8.3", nil)
		assert.NoError(t, err)

		err = validateRequest(r)
		assert.NoError(t, err)
	})

	t.Run("accept longer-than-semantic version", func(t *testing.T) {
		r, err := http.NewRequest("UPGRADE",
			"http://v1/outlet?node_id=000000&node_hostname=example.com&node_version=2.8.3.2-metapatch-edition", nil)
		assert.NoError(t, err)

		err = validateRequest(r)
		assert.NoError(t, err)
	})

	t.Run("reject too old version", func(t *testing.T) {
		r, err := http.NewRequest("UPGRADE",
			"http://v1/outlet?node_id=000000&node_hostname=example.com&node_version=2.4.82", nil)
		assert.NoError(t, err)

		err = validateRequest(r)
		assert.ErrorContains(t, err, "unsupported dataplane version")
	})

	t.Run("many builds does not a new version make", func(t *testing.T) {
		r, err := http.NewRequest("UPGRADE",
			"http://v1/outlet?node_id=000000&node_hostname=example.com&node_version=2.4.82.996-extra-strength", nil)
		assert.NoError(t, err)

		err = validateRequest(r)
		assert.ErrorContains(t, err, "unsupported dataplane version")
	})
}
