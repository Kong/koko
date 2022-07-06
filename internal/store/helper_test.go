package store

import (
	"testing"

	"github.com/kong/koko/internal/persistence"
	"github.com/stretchr/testify/assert"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

func Test_getPersistenceListOptions(t *testing.T) {
	expr := &exprpb.Expr{Id: 1}

	assert.Equal(
		t,
		&persistence.ListOpts{Limit: 10, Offset: 30, Filter: expr},
		getPersistenceListOptions(&ListOpts{PageSize: 10, Page: 4, Filter: expr}),
		"with pagination & filter",
	)
	assert.Equal(
		t,
		&persistence.ListOpts{Limit: 10, Offset: 0},
		getPersistenceListOptions(&ListOpts{PageSize: 10, Page: 0}),
		"first page (page=0)",
	)
	assert.Equal(
		t,
		&persistence.ListOpts{Limit: 10, Offset: 0},
		getPersistenceListOptions(&ListOpts{PageSize: 10, Page: 1}),
		"first page (page=1)",
	)
}
