package persistence

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPersister(t *testing.T) {
	memoryPersister := &Memory{}
	sqlitePersister, err := NewSQLite("file::memory:?cache=shared")
	assert.Nil(t, err)

	persisters := map[string]Persister{
		"memory": memoryPersister,
		"sqlite": sqlitePersister,
	}
	for name, p := range persisters {
		t.Run(name, func(t *testing.T) {
			// put
			value := []byte("foo-value")
			assert.Nil(t, p.Put(context.Background(), "foo", value))
			// get
			gotValue, err := p.Get(context.Background(), "foo")
			assert.Nil(t, err)
			assert.Equal(t, value, gotValue)

			// delete
			assert.Nil(t, p.Delete(context.Background(), "foo"))
			// get after delete
			gotValue, err = p.Get(context.Background(), "foo")
			assert.ErrorAs(t, err, &ErrNotFound{})
			assert.Nil(t, gotValue)
			// list
			bazValue := []byte("baz-value")
			assert.Nil(t, p.Put(context.Background(), "baz", bazValue))
			barValue := []byte("bar-value")
			assert.Nil(t, p.Put(context.Background(), "bar", barValue))
			gotValues, err := p.List(context.Background(), "ba")
			assert.Nil(t, err)
			assert.ElementsMatch(t, [][]byte{bazValue, barValue}, gotValues)
		})
	}
}
