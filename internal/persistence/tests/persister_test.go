package tests

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/kong/koko/internal/persistence"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
)

func TestPersister(t *testing.T) {
	p, err := util.GetPersister()
	require.Nil(t, err)

	t.Run("Get()", func(t *testing.T) {
		t.Run("returns an existing value", func(t *testing.T) {
			// put
			value := []byte("value1")
			require.Nil(t, p.Put(context.Background(), "key1", value))
			// get
			gotValue, err := p.Get(context.Background(), "key1")
			require.Nil(t, err)
			require.Equal(t, value, gotValue)
		})
		t.Run("returns ErrNotFound when value doesn't exist", func(t *testing.T) {
			gotValue, err := p.Get(context.Background(),
				"key-does-not-exist")
			require.Equal(t, persistence.ErrNotFound{Key: "key-does-not-exist"}, err)
			require.Nil(t, gotValue)
		})
	})
	t.Run("Put()", func(t *testing.T) {
		t.Run("stores a value", func(t *testing.T) {
			// put
			value := []byte("value2")
			require.Nil(t, p.Put(context.Background(), "key2", value))
			// get
			gotValue, err := p.Get(context.Background(), "key2")
			require.Nil(t, err)
			require.Equal(t, value, gotValue)
		})
		t.Run("overwrites a value", func(t *testing.T) {
			// put
			value := []byte("value3")
			require.Nil(t, p.Put(context.Background(), "key3", value))
			// get
			gotValue, err := p.Get(context.Background(), "key3")
			require.Nil(t, err)
			require.Equal(t, value, gotValue)

			value = []byte("value3-new")
			require.Nil(t, p.Put(context.Background(), "key3", value))
			// get
			gotValue, err = p.Get(context.Background(), "key3")
			require.Nil(t, err)
			require.Equal(t, value, gotValue)
		})
	})
	t.Run("Delete()", func(t *testing.T) {
		t.Run("deletes when key exist", func(t *testing.T) {
			value := []byte("value4")
			require.Nil(t, p.Put(context.Background(), "key4", value))
			require.Nil(t, p.Delete(context.Background(), "key4"))
			// get after delete
			gotValue, err := p.Get(context.Background(), "key4")
			require.ErrorAs(t, err,
				&persistence.ErrNotFound{Key: "key4"})
			require.Nil(t, gotValue)
		})
		t.Run("deletes fails when a key does not exist",
			func(t *testing.T) {
				require.Equal(t,
					persistence.ErrNotFound{Key: "key-no-exist"}, p.Delete(
						context.Background(), "key-no-exist"))
			})
	})
	t.Run("List()", func(t *testing.T) {
		t.Run("lists all keys with prefix", func(t *testing.T) {
			var expectedValues, expectedKeys []string
			for i := 0; i < 1000; i++ {
				value := []byte(fmt.Sprintf("prefix-value-%d", i))
				key := fmt.Sprintf("prefix/key%d", i)
				require.Nil(t, p.Put(context.Background(), key, value))
				expectedKeys = append(expectedKeys, key)
				expectedValues = append(expectedValues, string(value))
			}
			kvs, err := p.List(context.Background(), "prefix/")
			require.Nil(t, err)
			require.Len(t, kvs, 1000)

			var valuesAsStrings []string
			var keysAsStrings []string
			for _, kv := range kvs {
				key := string(kv[0])
				value := string(kv[1])
				keysAsStrings = append(keysAsStrings, key)
				valuesAsStrings = append(valuesAsStrings, value)
			}
			sort.Strings(keysAsStrings)
			sort.Strings(expectedKeys)
			sort.Strings(valuesAsStrings)
			sort.Strings(expectedValues)

			require.Equal(t, expectedKeys, keysAsStrings)
			require.Equal(t, expectedValues, valuesAsStrings)
		})
		t.Run("other prefixes are left as is", func(t *testing.T) {
			var expected []string
			for i := 0; i < 1000; i++ {
				value := []byte(fmt.Sprintf("prefix-value-%d", i))
				require.Nil(t, p.Put(
					context.Background(),
					fmt.Sprintf("prefix/key%d", i),
					value))
				expected = append(expected, string(value))
				// other prefixes
				value = []byte(fmt.Sprintf("other-prefix-value-%d", i))
				require.Nil(t, p.Put(
					context.Background(),
					fmt.Sprintf("ix/prefix/key%d", i),
					value))
			}
			values, err := p.List(context.Background(), "prefix/")
			require.Nil(t, err)
			require.Len(t, values, 1000)
			var valuesAsStrings []string
			for _, value := range values {
				valuesAsStrings = append(valuesAsStrings, string(value[1]))
			}
			sort.Strings(valuesAsStrings)
			sort.Strings(expected)
			require.Equal(t, expected, valuesAsStrings)
		})
	})
	t.Run("Tx()", func(t *testing.T) {
		t.Run("transaction rollbacks correctly", func(t *testing.T) {
			ctx := context.Background()
			tx, err := p.Tx(ctx)
			require.Nil(t, err)
			err = tx.Put(ctx, "key5", []byte("value5"))
			require.Nil(t, err)
			value, err := tx.Get(ctx, "key5")
			require.Nil(t, err)
			require.Equal(t, []byte("value5"), value)
			err = tx.Rollback()
			require.Nil(t, err)
			value, err = p.Get(ctx, "key5")
			require.Equal(t, persistence.ErrNotFound{Key: "key5"}, err)
			require.Nil(t, value)
		})
	})
}
