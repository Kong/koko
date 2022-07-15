package store

import (
	"fmt"
	"testing"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type testObjWithResourceDTO struct {
	model.Object
	data map[string]interface{}
}

func (t testObjWithResourceDTO) MarshalResourceJSON() ([]byte, error) {
	return json.Marshal(t.data)
}

func (t testObjWithResourceDTO) UnmarshalResourceJSON(b []byte) error {
	return json.Unmarshal(b, &t.data)
}

func Test_wrapObject(t *testing.T) {
	t.Run("marshal with default Protobuf marshaller", func(t *testing.T) {
		obj := resource.Consumer{Consumer: &v1.Consumer{Username: "test"}}
		objJSON, err := wrapObject(obj)
		require.NoError(t, err)
		assert.JSONEq(
			t,
			fmt.Sprintf(`{"type": %d, "object": {"username": "test"}}`, valueTypeObject),
			string(objJSON),
		)
	})

	t.Run("marshal with MarshalResourceJSON()", func(t *testing.T) {
		obj := testObjWithResourceDTO{
			data: map[string]interface{}{"key": "value"},
		}
		objJSON, err := wrapObject(obj)
		require.NoError(t, err)
		assert.JSONEq(
			t,
			fmt.Sprintf(`{"type": %d, "object": {"key": "value"}}`, valueTypeObject),
			string(objJSON),
		)
	})
}

func Test_unwrapObject(t *testing.T) {
	t.Run("unmarshal with default Protobuf unmarshaller", func(t *testing.T) {
		obj := resource.NewConsumer()
		require.NoError(t, unwrapObject(
			[]byte(fmt.Sprintf(`{"type": %d, "object": {"username": "test"}}`, valueTypeObject)),
			obj,
		))
		assert.True(t, proto.Equal(&v1.Consumer{Username: "test"}, obj.Consumer))
	})

	t.Run("unmarshal with UnmarshalResourceJSON()", func(t *testing.T) {
		obj := testObjWithResourceDTO{data: map[string]interface{}{}}
		require.NoError(t, unwrapObject(
			[]byte(fmt.Sprintf(`{"type": %d, "object": {"key": "value"}}`, valueTypeObject)),
			obj,
		))
		assert.Equal(t, map[string]interface{}{"key": "value"}, obj.data)
	})
}
