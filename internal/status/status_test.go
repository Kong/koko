package status

import "testing"

func TestMessageForCode(t *testing.T) {
	type args struct {
		code    Code
		message string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "string format check",
			args: args{
				code:    DPMissingPlugin,
				message: "foo, bar",
			},
			want: "kong data-plane node missing plugin[DP001]: foo, bar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MessageForCode(tt.args.code, tt.args.message); got != tt.want {
				t.Errorf("MessageForCode() = %v, want %v", got, tt.want)
			}
		})
	}
}
