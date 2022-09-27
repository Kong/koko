package db

import (
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_driverForDialect(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)

	t.Run("Check for unimplemented dialects", func(t *testing.T) {
		for _, dialect := range Dialects {
			if _, err := driverForDialect(dialect, db); err != nil {
				if strings.HasPrefix(err.Error(), "unsupported database") {
					assert.NoError(t, err)
				}
			}
		}
	})
}
