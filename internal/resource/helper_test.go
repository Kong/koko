package resource

import (
	"testing"

	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func validUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

func TestSetResource(t *testing.T) {
	tests := []struct {
		name        string
		object      model.Object
		resource    model.Resource
		expected    model.Object
		expectedErr string
	}{
		{
			name:     "non-matching descriptors",
			object:   &Consumer{},
			resource: &v1.Target{},
			expectedErr: `unable to set resource: expected "kong.admin.model.v1.Consumer" ` +
				`but got "kong.admin.model.v1.Target"`,
		},
		{
			name:        "nil resource on object",
			object:      &Consumer{},
			resource:    &v1.Consumer{},
			expectedErr: `unable to set resource: got invalid destination resource`,
		},
		{
			name:     "successfully set resource",
			object:   &Consumer{Consumer: &v1.Consumer{Id: "id1"}},
			resource: &v1.Consumer{Id: "id2"},
			expected: &Consumer{Consumer: &v1.Consumer{Id: "id2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetResource(tt.object, tt.resource)
			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)
			assert.True(t, proto.Equal(tt.expected.Resource(), tt.resource))
		})
	}
}
