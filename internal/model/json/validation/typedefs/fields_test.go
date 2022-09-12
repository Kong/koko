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

func TestPathPattern(t *testing.T) {
	// regex associated with prefix paths and the 3.0 regex path format.
	pathRegex := regexp.MustCompile(RouterPath.AllOf[0].Pattern)

	tests := []struct {
		name     string
		paths    []string
		expected bool
	}{
		{
			name:     "must not be empty",
			paths:    []string{""},
			expected: false,
		},
		{
			name: "must start with / for prefix paths",
			paths: []string{
				"/",
				"/foo",
				"/kong",
				"/koko",
				"/insomnia",
				"/abcd~user~2",
				"/abcd%aa%10%ff%AA%FF",
				"/koko/",
			},
			expected: true,
		},
		{
			name: "must start with ~/ for regex paths",
			paths: []string{
				"~/.*",
				"~/[fF][oO]{2}",
				"~/kong|koko|insomnia",
			},
			expected: true,
		},
		{
			name: "fails for invalid paths",
			paths: []string{
				"     ",
				"~kong",
				"koko",
				"#$%",
				"%20",
			},
			expected: false,
		},
	}

	for _, test := range tests {
		for _, path := range test.paths {
			valid := pathRegex.MatchString(path)
			require.Equal(t, test.expected, valid, "test: '%v', path: '%v'", test.name, path)
		}
	}
}
