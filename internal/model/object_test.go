package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultiValueIndex(t *testing.T) {
	type args struct {
		values []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no args",
			args: args{
				values: nil,
			},
			want: "",
		},
		{
			name: "one value",
			args: args{
				values: []string{"foo"},
			},
			want: "foo",
		},
		{
			name: "two values",
			args: args{
				values: []string{"foo", "bar"},
			},
			want: "foo:bar",
		},
		{
			name: "more than two values",
			args: args{
				values: []string{"foo", "bar", "baz", "fuz"},
			},
			want: "foo:bar:baz:fuz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MultiValueIndex(tt.args.values...); got != tt.want {
				t.Errorf("MultiValueIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIndex_Validate(t *testing.T) {
	tests := []struct {
		name          string
		index         Index
		expectedError string
	}{
		{
			name: "invalid name",
			index: Index{
				Type:  IndexUnique,
				Value: "something",
			},
			expectedError: "index name is not set",
		},
		{
			name: "invalid type",
			index: Index{
				Name:  "something",
				Value: "something",
			},
			expectedError: "index type is not set",
		},
		{
			name: "invalid value",
			index: Index{
				Name: "something",
				Type: IndexUnique,
			},
			expectedError: "index value is not set",
		},
		{
			name: "invalid field name: prefix provided",
			index: Index{
				FieldName: "$.something",
				Name:      "something",
				Type:      IndexUnique,
				Value:     "something",
			},
			expectedError: `must not include JSONPath prefix ("$.") in field name`,
		},
		{
			name: "invalid field name: invalid JSONPath",
			index: Index{
				FieldName: "something.",
				Name:      "something",
				Type:      IndexUnique,
				Value:     "something",
			},
			expectedError: `invalid JSONPath field name: expected JSON child identifier after '.' at 13`,
		},
		{
			name: "invalid foreign type: not provided",
			index: Index{
				Name:  "something",
				Type:  IndexForeign,
				Value: "something",
			},
			expectedError: `index foreign type is not set`,
		},
		{
			name: "valid index object",
			index: Index{
				Name:        "something",
				Type:        IndexForeign,
				Value:       "something",
				ForeignType: "something",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.index.Validate()
			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestIndexes_Validate(t *testing.T) {
	tests := []struct {
		name          string
		indexes       Indexes
		expectedError string
	}{
		{
			name: "ensure regular validation executes: without name",
			indexes: Indexes{{
				Type:  IndexUnique,
				Value: "something",
			}},
			expectedError: "invalid index in list: index name is not set",
		},
		{
			name: "ensure regular validation executes: with name",
			indexes: Indexes{{
				Name: "something",
				Type: IndexUnique,
			}},
			expectedError: `invalid index (name = "something") in list: index value is not set`,
		},
		{
			name: "ensure regular validation executes: with field name",
			indexes: Indexes{{
				FieldName: "something",
				Type:      IndexUnique,
			}},
			expectedError: `invalid index (field_name = "something") in list: index value is not set`,
		},
		{
			name: "ensure no duplicates",
			indexes: Indexes{
				{
					ForeignType: "something",
					Name:        "something",
					Type:        IndexForeign,
					Value:       "something",
				},
				{
					ForeignType: "something",
					Name:        "something",
					Type:        IndexForeign,
					Value:       "something",
				},
			},
			expectedError: `invalid index (name = "something") in list: duplicate index contents`,
		},
		{
			name: "valid indexes",
			indexes: Indexes{
				{
					ForeignType: "something",
					Name:        "something",
					Type:        IndexForeign,
					Value:       "something",
				},
				{
					ForeignType: "something",
					Name:        "something",
					Type:        IndexForeign,
					Value:       "another thing",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.indexes.Validate()
			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
				return
			}
			assert.NoError(t, err)
		})
	}
}
