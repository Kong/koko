package db

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewPersister(t *testing.T) {
	t.Run("Check for unimplemented dialects", func(t *testing.T) {
		config := Config{Logger: zap.L()}
		for _, dialect := range Dialects {
			config.Dialect = dialect
			if _, err := NewPersister(config); err != nil {
				if strings.HasPrefix(err.Error(), "unsupported database") {
					assert.NoError(t, err)
				}
			}
		}
	})
}
