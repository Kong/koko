package persistence

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPersister(t *testing.T) {
	sqlitePersister, err := NewMemory()
	assert.Nil(t, err)

	persisters := map[string]Persister{
		"sqlite": sqlitePersister,
	}
	for name, p := range persisters {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			t.Run("Get()", func(t *testing.T) {
				t.Run("returns an existing value", func(t *testing.T) {
					// put
					value := []byte("value1")
					assert.Nil(t, p.Put(context.Background(), "key1", value))
					// get
					gotValue, err := p.Get(context.Background(), "key1")
					assert.Nil(t, err)
					assert.Equal(t, value, gotValue)
				})
				t.Run("returns ErrNotFound when value doesn't exist", func(t *testing.T) {
					gotValue, err := p.Get(context.Background(),
						"key-does-not-exist")
					assert.Equal(t, ErrNotFound{key: "key-does-not-exist"}, err)
					assert.Nil(t, gotValue)
				})
			})
			t.Run("Put()", func(t *testing.T) {
				t.Run("stores a value", func(t *testing.T) {
					// put
					value := []byte("value2")
					assert.Nil(t, p.Put(context.Background(), "key2", value))
					// get
					gotValue, err := p.Get(context.Background(), "key2")
					assert.Nil(t, err)
					assert.Equal(t, value, gotValue)
				})
				t.Run("overwrites a value", func(t *testing.T) {
					// put
					value := []byte("value3")
					assert.Nil(t, p.Put(context.Background(), "key3", value))
					// get
					gotValue, err := p.Get(context.Background(), "key3")
					assert.Nil(t, err)
					assert.Equal(t, value, gotValue)

					value = []byte("value3-new")
					assert.Nil(t, p.Put(context.Background(), "key3", value))
					// get
					gotValue, err = p.Get(context.Background(), "key3")
					assert.Nil(t, err)
					assert.Equal(t, value, gotValue)
				})
			})
			t.Run("Delete()", func(t *testing.T) {
				t.Run("deletes when key exist", func(t *testing.T) {
					value := []byte("value4")
					assert.Nil(t, p.Put(context.Background(), "key4", value))
					assert.Nil(t, p.Delete(context.Background(), "key4"))
					// get after delete
					gotValue, err := p.Get(context.Background(), "key4")
					assert.ErrorAs(t, err, &ErrNotFound{key: "key4"})
					assert.Nil(t, gotValue)
				})
				t.Run("deletes fails when a key does not exist",
					func(t *testing.T) {
						assert.Equal(t, ErrNotFound{key: "key-no-exist"}, p.Delete(
							context.Background(), "key-no-exist"))
					})
			})
			t.Run("List()", func(t *testing.T) {
				t.Run("lists all keys with prefix", func(t *testing.T) {
					var expected []string
					for i := 0; i < 1000; i++ {
						value := []byte(fmt.Sprintf("prefix-value-%d", i))
						assert.Nil(t, p.Put(
							context.Background(),
							fmt.Sprintf("prefix/key%d", i),
							value))
						expected = append(expected, string(value))
					}
					values, err := p.List(context.Background(), "prefix/")
					assert.Nil(t, err)
					assert.Len(t, values, 1000)
					var valuesAsStrings []string
					for _, value := range values {
						valuesAsStrings = append(valuesAsStrings, string(value))
					}
					sort.Strings(valuesAsStrings)
					sort.Strings(expected)
					assert.Equal(t, expected, valuesAsStrings)
				})
				t.Run("other prefixes are left as is", func(t *testing.T) {
					var expected []string
					for i := 0; i < 1000; i++ {
						value := []byte(fmt.Sprintf("prefix-value-%d", i))
						assert.Nil(t, p.Put(
							context.Background(),
							fmt.Sprintf("prefix/key%d", i),
							value))
						expected = append(expected, string(value))
						// other prefixes
						value = []byte(fmt.Sprintf("other-prefix-value-%d", i))
						assert.Nil(t, p.Put(
							context.Background(),
							fmt.Sprintf("ix/prefix/key%d", i),
							value))
					}
					values, err := p.List(context.Background(), "prefix/")
					assert.Nil(t, err)
					assert.Len(t, values, 1000)
					var valuesAsStrings []string
					for _, value := range values {
						valuesAsStrings = append(valuesAsStrings, string(value))
					}
					sort.Strings(valuesAsStrings)
					sort.Strings(expected)
					assert.Equal(t, expected, valuesAsStrings)
				})
			})
			t.Run("Tx()", func(t *testing.T) {
				t.Run("transaction rollbacks correctly", func(t *testing.T) {
					ctx := context.Background()
					tx, err := p.Tx(ctx)
					assert.Nil(t, err)
					err = tx.Put(ctx, "key5", []byte("value5"))
					assert.Nil(t, err)
					value, err := tx.Get(ctx, "key5")
					assert.Nil(t, err)
					assert.Equal(t, []byte("value5"), value)
					err = tx.Rollback()
					assert.Nil(t, err)
					value, err = p.Get(ctx, "key5")
					assert.Equal(t, ErrNotFound{"key5"}, err)
					assert.Nil(t, value)
				})
			})
		})
	}
}

func TestErrNotFound(t *testing.T) {
	err := ErrNotFound{key: "foo"}
	assert.Equal(t, "foo", err.Key())
	assert.Equal(t, "foo not found", err.Error())
}
