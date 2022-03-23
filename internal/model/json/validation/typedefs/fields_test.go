package typedefs

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNamePattern(t *testing.T) {
	require.NotPanics(t, func() {
		re := regexp.MustCompile(namePattern)
		require.True(t, re.MatchString("foo-bar"))
	})
}

func TestTagPattern(t *testing.T) {
	t.Run("colon is acceptable in tag", func(t *testing.T) {
		re := regexp.MustCompile(tagPattern)
		require.True(t, re.MatchString("foo:bar"))
	})
}
