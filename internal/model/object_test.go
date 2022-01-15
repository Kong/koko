package model

import "testing"

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
