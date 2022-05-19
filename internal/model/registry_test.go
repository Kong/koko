package model

import (
	"testing"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/proto"
)

type testConsumerObj struct {
	Object
	*v1.Consumer
}

func (o *testConsumerObj) Resource() Resource { return o.Consumer }
func (o *testConsumerObj) SetResource(r Resource) error {
	proto.Merge(o.Consumer, r)
	return nil
}

type registerTypeSuite struct{ suite.Suite }

func (s *registerTypeSuite) AfterTest(string, string) { resetTypes() }

func (s *registerTypeSuite) TestAlreadyRegistered() {
	require.NoError(s.T(), RegisterType("foo", &v1.Consumer{}, nil))
	assert.EqualError(s.T(), RegisterType("foo", &v1.Consumer{}, nil), "type already registered: foo")
}

func (s *registerTypeSuite) TestEmptyProtoMessage() {
	assert.EqualError(s.T(), RegisterType("foo", nil, nil), "must not provide empty Protobuf message")
}

func (s *registerTypeSuite) TestInvalidProtoMessage() {
	assert.EqualError(s.T(), RegisterType("foo", (*v1.Consumer)(nil), nil), "must not provide invalid Protobuf message")
}

func (s *registerTypeSuite) TestProtoMessageAlreadyRegistered() {
	require.NoError(s.T(), RegisterType("foo", &v1.Consumer{}, nil))
	assert.EqualError(
		s.T(),
		RegisterType("bar", &v1.Consumer{}, nil),
		"protobuf message already registered: kong.admin.model.v1.Consumer",
	)
}

func (s *registerTypeSuite) TestSuccessfullyRegisteredType() {
	typ, r, p := Type("foo"), &testConsumerObj{}, &v1.Consumer{}
	require.NoError(s.T(), RegisterType(typ, p, func() Object { return r }))
	assert.Equal(s.T(), r, types[typ]())
	assert.Equal(s.T(), typ, protoToType["kong.admin.model.v1.Consumer"])
}

func TestRegisterType(t *testing.T) {
	suite.Run(t, &registerTypeSuite{})
}

func TestAllTypes(t *testing.T) {
	defer resetTypes()

	require.Nil(t, RegisterType("foo", &v1.Consumer{}, func() Object {
		return nil
	}))
	require.Nil(t, RegisterType("bar", &v1.Route{}, func() Object {
		return nil
	}))
	require.Nil(t, RegisterType("baz", &v1.Certificate{}, func() Object {
		return nil
	}))
	require.ElementsMatch(t, []Type{"foo", "bar", "baz"}, AllTypes())
}

func TestObjectFromProto(t *testing.T) {
	defer resetTypes()

	require.NoError(t, RegisterType("test_object", &v1.Consumer{}, func() Object {
		return &testConsumerObj{Consumer: &v1.Consumer{}}
	}))

	tests := []struct {
		name        string
		proto       proto.Message
		expected    Object
		expectedErr string
	}{
		{
			name:        "nil proto",
			expectedErr: "cannot resolve empty Protobuf message to object",
		},
		{
			name:        "invalid proto",
			proto:       (*v1.ActiveHealthcheck)(nil),
			expectedErr: "cannot resolve invalid Protobuf message to object",
		},
		{
			name:        "unknown type",
			proto:       &v1.ActiveHealthcheck{},
			expectedErr: "cannot find type from Protobuf message kong.admin.model.v1.ActiveHealthcheck",
		},
		{
			name:     "resolve Protobuf message to object successfully",
			proto:    &v1.Consumer{Id: "test-id"},
			expected: &testConsumerObj{Consumer: &v1.Consumer{Id: "test-id"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := ObjectFromProto(tt.proto)
			if tt.expectedErr != "" {
				require.EqualError(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)
			assert.True(t, proto.Equal(tt.expected.Resource(), actual.Resource()))
		})
	}
}

func resetTypes() {
	types, protoToType = map[Type]func() Object{}, map[string]Type{}
}
