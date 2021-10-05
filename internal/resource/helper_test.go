package resource

import (
	"testing"

	"github.com/kong/koko/internal/model/validation"
	"github.com/kong/koko/internal/model/validation/typedefs"
	"github.com/stretchr/testify/assert"
)

func Test_notHTTPProtocol(t *testing.T) {
	type args struct {
		protocol string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "http is valid",
			args: args{
				protocol: "http",
			},
			want: false,
		},
		{
			name: "https is valid",
			args: args{
				protocol: "https",
			},
			want: false,
		},
		{
			name: "grpc is invalid",
			args: args{
				protocol: "grpc",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := notHTTPProtocol(tt.args.protocol); got != tt.want {
				t.Errorf("notHTTPProtocol() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeRules(t *testing.T) {
	// merge rule arrays
	rules := mergeRules(typedefs.UUID(), typedefs.Protocol())
	assert.NotNil(t, rules)

	assert.Panics(t, func() {
		mergeRules("foo")
	})
}

func fieldsFromErr(err validation.Error) []string {
	res := make([]string, 0, len(err.Fields))
	for _, f := range err.Fields {
		res = append(res, f.Name)
	}
	return res
}
